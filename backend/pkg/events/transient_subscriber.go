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

// TransientSubscriberConfig holds the configuration options for initializing an TransientSubscriber.
type TransientSubscriberConfig struct {
	BrokerURL      string
	Logger         *slog.Logger
	ExchangeName   string
	Keys           []string
	InitialCredits int32
	CloseTimeout   time.Duration
}

// TransientSubscriber consumes messages from a RabbitMQ exchange and delivers them via the Deliveries channel.
type TransientSubscriber struct {
	cfg    *TransientSubscriberConfig
	logger *slog.Logger

	conn     *rmq.AmqpConnection
	env      *rmq.Environment
	consumer *rmq.Consumer

	mu        sync.RWMutex
	available bool
	wg        sync.WaitGroup

	consumeCancel context.CancelFunc

	Deliveries chan *ReceivedMessage
}

// NewTransientSubscriber creates and returns a new instance of TransientSubscriber.
func NewTransientSubscriber(cfg *TransientSubscriberConfig) *TransientSubscriber {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.CloseTimeout == 0 {
		cfg.CloseTimeout = DefaultCloseTimeout
	}
	if cfg.InitialCredits == 0 {
		cfg.InitialCredits = DefaultInitialCredits
	}

	return &TransientSubscriber{
		cfg:           cfg,
		logger:        cfg.Logger,
		conn:          nil,
		env:           rmq.NewEnvironment(cfg.BrokerURL, nil),
		consumer:      nil,
		mu:            sync.RWMutex{},
		available:     false,
		wg:            sync.WaitGroup{},
		consumeCancel: nil,
		Deliveries:    nil,
	}
}

// Start establishes the connection to the broker, declares the exchange and an exclusive queue,
// binds the specified routing keys, and starts the consume loop.
func (c *TransientSubscriber) Start(ctx context.Context) error {
	conn, err := amqp.ConnectWithTopicExchange(ctx, c.env, c.cfg.ExchangeName)
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

	if err := amqp.BindExchangeToQueue(ctx, conn, c.cfg.ExchangeName, qInfo.Name(), c.cfg.Keys); err != nil {
		c.closeConn(ctx)
		c.logger.Error("Failed to bind exchange to queue.")
		return fmt.Errorf("bind exchange to queue: %w", err)
	}

	consumer, err := conn.NewConsumer(ctx, qInfo.Name(), &rmq.ConsumerOptions{
		InitialCredits: c.cfg.InitialCredits,
	})
	if err != nil {
		c.closeConn(ctx)
		c.cfg.Logger.Error("Failed to create consumer", "error", err)
		return fmt.Errorf("create consumer: %w", err)
	}
	c.consumer = consumer

	consumeCtx, consumeCancel := context.WithCancel(ctx)
	c.consumeCancel = consumeCancel

	c.Deliveries = make(chan *ReceivedMessage, c.cfg.InitialCredits)
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

	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.CloseTimeout)
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

func (c *TransientSubscriber) closeConn(ctx context.Context) {
	if err := c.env.CloseConnections(ctx); err != nil {
		c.logger.Error("Failed to close connections", "error", err)
	}
}
