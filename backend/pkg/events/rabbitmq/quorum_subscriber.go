package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	goamqp "github.com/Azure/go-amqp"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"

	"github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq/amqp"
)

// QuorumSubscriberOption configures a QuorumSubscriber.
type QuorumSubscriberOption func(*QuorumSubscriber)

// QuorumSubscriber consumes messages from a RabbitMQ exchange and delivers them via the Deliveries channel.
type QuorumSubscriber struct {
	env          *rmq.Environment
	exchangeName string
	queueName    string

	keys           []string
	initialCredits int32
	closeTimeout   time.Duration
	logger         *slog.Logger

	conn     *rmq.AmqpConnection
	consumer *rmq.Consumer

	mu        sync.RWMutex
	available bool
	wg        sync.WaitGroup

	consumeCancel context.CancelFunc

	deliveries chan *ReceivedMessage
}

// NewQuorumSubscriber creates and returns a new instance of QuorumSubscriber.
func NewQuorumSubscriber(
	env *rmq.Environment, exchangeName, queueName string, logger *slog.Logger, opts ...QuorumSubscriberOption,
) *QuorumSubscriber {
	logger = logger.With(
		slog.String("module", "events/rabbitmq"),
		slog.String("component", "quorum_subscriber"),
	)

	subscriber := &QuorumSubscriber{
		env:            env,
		exchangeName:   exchangeName,
		queueName:      queueName,
		keys:           nil,
		initialCredits: DefaultInitialCredits,
		closeTimeout:   DefaultCloseTimeout,
		logger:         logger,
		conn:           nil,
		consumer:       nil,
		mu:             sync.RWMutex{},
		available:      false,
		wg:             sync.WaitGroup{},
		consumeCancel:  nil,
		deliveries:     nil,
	}

	for _, opt := range opts {
		opt(subscriber)
	}

	return subscriber
}

// Start establishes the connection to the broker, declares the exchange and an exclusive queue,
// binds the specified routing keys, and starts the consume loop.
func (s *QuorumSubscriber) Start(ctx context.Context) (<-chan *ReceivedMessage, error) {
	s.mu.Lock()
	if s.available {
		s.mu.Unlock()
		return nil, ErrSubscriberAlreadyStarted
	}
	s.mu.Unlock()

	conn, err := amqp.ConnectWithTopicExchange(ctx, s.env, s.exchangeName)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to connect to topic exchange", slog.Any("error", err))
		return nil, fmt.Errorf("connect to topic exchange: %w", err)
	}

	qInfo, declareErr := amqp.DeclareQuorumQueue(ctx, conn, s.queueName)
	if declareErr != nil {
		declareErr = fmt.Errorf("declare quorum queue: %w", declareErr)

		if closeErr := conn.Close(ctx); closeErr != nil {
			return nil, errors.Join(declareErr, fmt.Errorf("close AMQP connection: %w", closeErr))
		}

		return nil, declareErr
	}

	if err := amqp.BindExchangeToQueue(ctx, conn, s.exchangeName, qInfo.Name(), s.keys); err != nil {
		s.logger.ErrorContext(ctx, "failed to bind exchange to queue", slog.Any("error", err))
		err = fmt.Errorf("bind exchange to queue: %w", err)

		if closeErr := conn.Close(ctx); closeErr != nil {
			return nil, errors.Join(err, fmt.Errorf("close AMQP connection: %w", closeErr))
		}

		return nil, err
	}

	consumer, err := conn.NewConsumer(ctx, qInfo.Name(), &rmq.ConsumerOptions{
		InitialCredits: s.initialCredits,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create consumer", slog.Any("error", err))
		err = fmt.Errorf("create consumer: %w", err)

		if closeErr := conn.Close(ctx); closeErr != nil {
			return nil, errors.Join(err, fmt.Errorf("close AMQP connection: %w", closeErr))
		}

		return nil, err
	}

	s.consumer = consumer

	consumeCtx, consumeCancel := context.WithCancel(ctx)
	s.consumeCancel = consumeCancel

	s.deliveries = make(chan *ReceivedMessage, s.initialCredits)

	s.wg.Add(1)
	go s.consumeLoop(consumeCtx)

	s.mu.Lock()
	s.available = true
	s.mu.Unlock()

	return s.deliveries, nil
}

// Close gracefully shuts down the consumer, cancels the consume loop, waits for in-flight
// messages, and closes the AMQP connections.
func (s *QuorumSubscriber) Close(ctx context.Context) error {
	s.mu.Lock()
	s.available = false
	s.mu.Unlock()

	if s.consumeCancel != nil {
		s.consumeCancel()
	}

	s.wg.Wait()
	s.logger.DebugContext(ctx, "consumer loop stopped gracefully")

	ctx, cancel := context.WithTimeout(ctx, s.closeTimeout)
	defer cancel()

	var errs []error

	if s.consumer != nil {
		if err := s.consumer.Close(ctx); err != nil {
			s.logger.ErrorContext(ctx, "failed to close consumer", slog.Any("error", err))
			errs = append(errs, err)
		}
	}

	if err := s.env.CloseConnections(ctx); err != nil {
		s.logger.ErrorContext(ctx, "failed to close connections", slog.Any("error", err))
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (s *QuorumSubscriber) consumeLoop(ctx context.Context) {
	defer s.wg.Done()
	defer close(s.deliveries)

	for {
		delivery, err := s.consumer.Receive(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				s.logger.DebugContext(ctx, "consumer loop exiting due to shutdown signal")
				return
			}

			s.logger.ErrorContext(ctx, "failed to receive message, consumer exiting", slog.Any("error", err))

			return
		}

		msg := delivery.Message()

		var data []byte
		if len(msg.Data) > 0 {
			data = msg.Data[0]
		}

		received := &ReceivedMessage{
			Data:        data,
			acceptFunc:  func() error { return delivery.Accept(ctx) },
			requeueFunc: func() error { return delivery.Requeue(ctx) },
			discardFunc: func(e *goamqp.Error) error { return delivery.Discard(ctx, e) },
		}

		select {
		case s.deliveries <- received:
		case <-ctx.Done():
			if err := received.Requeue(); err != nil {
				s.logger.ErrorContext(ctx, "failed to nack message", slog.Any("error", err))
			}

			return
		}
	}
}

// WithQuorumSubscriberKeys sets the routing keys for the subscriber to bind to.
func WithQuorumSubscriberKeys(keys []string) QuorumSubscriberOption {
	return func(s *QuorumSubscriber) {
		s.keys = keys
	}
}

// WithQuorumSubscriberInitialCredits sets the initial credit (prefetch) count for the consumer.
func WithQuorumSubscriberInitialCredits(initialCredits int32) QuorumSubscriberOption {
	return func(s *QuorumSubscriber) {
		s.initialCredits = initialCredits
	}
}

// WithQuorumSubscriberCloseTimeout sets the close timeout duration for the QuorumSubscriber.
func WithQuorumSubscriberCloseTimeout(timeout time.Duration) QuorumSubscriberOption {
	return func(s *QuorumSubscriber) {
		s.closeTimeout = timeout
	}
}
