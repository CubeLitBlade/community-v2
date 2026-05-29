package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	goamqp "github.com/Azure/go-amqp"
	"github.com/cubelitblade/community-v2/backend/pkg/events/internal/amqp"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

// TransientSubscriberOption configures a TransientSubscriber.
type TransientSubscriberOption func(*TransientSubscriber)

// TransientSubscriber consumes messages from a RabbitMQ exchange and delivers them via the Deliveries channel.
type TransientSubscriber struct {
	env          *rmq.Environment
	exchangeName string

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

	Deliveries chan *ReceivedMessage
}

// NewTransientSubscriber creates and returns a new instance of TransientSubscriber.
func NewTransientSubscriber(
	env *rmq.Environment, exchangeName string, opts ...TransientSubscriberOption,
) *TransientSubscriber {
	subscriber := &TransientSubscriber{
		env:            env,
		exchangeName:   exchangeName,
		keys:           nil,
		initialCredits: DefaultInitialCredits,
		closeTimeout:   DefaultCloseTimeout,
		logger:         slog.Default(),
		conn:           nil,
		consumer:       nil,
		mu:             sync.RWMutex{},
		available:      false,
		wg:             sync.WaitGroup{},
		consumeCancel:  nil,
		Deliveries:     nil,
	}

	for _, opt := range opts {
		opt(subscriber)
	}

	return subscriber
}

// Start establishes the connection to the broker, declares the exchange and an exclusive queue,
// binds the specified routing keys, and starts the consume loop.
func (c *TransientSubscriber) Start(ctx context.Context) error {
	conn, err := amqp.ConnectWithTopicExchange(ctx, c.env, c.exchangeName)
	if err != nil {
		c.logger.Error("Failed to connect to topic exchange.", "err", err)
		return fmt.Errorf("connect to topic exchange: %w", err)
	}

	qInfo, declareErr := amqp.DeclareExclusiveQueue(ctx, conn)
	if declareErr != nil {
		declareErr = fmt.Errorf("declare exclusive queue: %w", declareErr)

		if closeErr := conn.Close(ctx); closeErr != nil {
			return errors.Join(declareErr, fmt.Errorf("close AMQP connection: %w", closeErr))
		}

		return declareErr
	}

	if err := amqp.BindExchangeToQueue(ctx, conn, c.exchangeName, qInfo.Name(), c.keys); err != nil {
		c.logger.Error("Failed to bind exchange to queue.", "err", err)
		err = fmt.Errorf("bind exchange to queue: %w", err)

		if closeErr := conn.Close(ctx); closeErr != nil {
			return errors.Join(err, fmt.Errorf("close AMQP connection: %w", closeErr))
		}

		return err
	}

	consumer, err := conn.NewConsumer(ctx, qInfo.Name(), &rmq.ConsumerOptions{
		InitialCredits: c.initialCredits,
	})
	if err != nil {
		c.logger.Error("Failed to create consumer", "err", err)
		err = fmt.Errorf("create consumer: %w", err)

		if closeErr := conn.Close(ctx); closeErr != nil {
			return errors.Join(err, fmt.Errorf("close AMQP connection: %w", closeErr))
		}

		return err
	}
	c.consumer = consumer

	consumeCtx, consumeCancel := context.WithCancel(ctx)
	c.consumeCancel = consumeCancel

	c.Deliveries = make(chan *ReceivedMessage, c.initialCredits)
	c.wg.Add(1)
	go c.consumeLoop(consumeCtx)

	c.mu.Lock()
	c.available = true
	c.mu.Unlock()

	return nil
}

// Close gracefully shuts down the consumer, cancels the consume loop, waits for in-flight
// messages, and closes the AMQP connections.
func (c *TransientSubscriber) Close() error {
	c.mu.Lock()
	c.available = false
	c.mu.Unlock()

	if c.consumeCancel != nil {
		c.consumeCancel()
	}

	c.wg.Wait()
	c.logger.Debug("Consumer loop stopped gracefully.")

	ctx, cancel := context.WithTimeout(context.Background(), c.closeTimeout)
	defer cancel()

	var errs []error

	if c.consumer != nil {
		if err := c.consumer.Close(ctx); err != nil {
			c.logger.Error("Failed to close consumer", "error", err)
			errs = append(errs, err)
		}
	}

	if err := c.env.CloseConnections(ctx); err != nil {
		c.logger.Error("Failed to close connections", "error", err)
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (c *TransientSubscriber) consumeLoop(ctx context.Context) {
	defer c.wg.Done()
	defer close(c.Deliveries)

	for {
		delivery, err := c.consumer.Receive(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				c.logger.Debug("Consumer loop exiting due to shutdown signal.")
				return
			}

			c.logger.Error("Failed to receive message, consumer exiting.", "error", err)
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
		case c.Deliveries <- received:
		case <-ctx.Done():
			if err := received.Requeue(); err != nil {
				c.logger.Error("Failed to nack message", "error", err)
			}
			return
		}
	}
}

// WithTransientSubscriberKeys sets the routing keys for the subscriber to bind to.
func WithTransientSubscriberKeys(keys []string) TransientSubscriberOption {
	return func(s *TransientSubscriber) {
		s.keys = keys
	}
}

// WithTransientSubscriberInitialCredits sets the initial credit (prefetch) count for the consumer.
func WithTransientSubscriberInitialCredits(initialCredits int32) TransientSubscriberOption {
	return func(s *TransientSubscriber) {
		s.initialCredits = initialCredits
	}
}

// WithTransientSubscriberLogger sets the logger for the TransientSubscriber.
func WithTransientSubscriberLogger(logger *slog.Logger) TransientSubscriberOption {
	return func(s *TransientSubscriber) {
		s.logger = logger
	}
}

// WithTransientSubscriberCloseTimeout sets the close timeout duration for the TransientSubscriber.
func WithTransientSubscriberCloseTimeout(timeout time.Duration) TransientSubscriberOption {
	return func(s *TransientSubscriber) {
		s.closeTimeout = timeout
	}
}
