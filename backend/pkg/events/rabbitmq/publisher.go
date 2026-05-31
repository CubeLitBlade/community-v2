package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq/amqp"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

// PublisherOption configures a Publisher.
type PublisherOption func(*Publisher)

// Publisher manages the publishing of messages to a specific RabbitMQ exchange.
type Publisher struct {
	env          *rmq.Environment
	exchangeName string

	closeTimeout time.Duration
	logger       *slog.Logger

	conn      *rmq.AmqpConnection
	publisher *rmq.Publisher

	mu        sync.RWMutex
	available bool
	wg        sync.WaitGroup
}

// NewPublisher creates and returns a new instance of Publisher.
func NewPublisher(
	env *rmq.Environment, exchangeName string, opts ...PublisherOption,
) *Publisher {
	publisher := &Publisher{
		env:          env,
		exchangeName: exchangeName,
		closeTimeout: DefaultCloseTimeout,
		logger:       slog.Default(),
		conn:         nil,
		publisher:    nil,
		mu:           sync.RWMutex{},
		available:    false,
		wg:           sync.WaitGroup{},
	}

	for _, opt := range opts {
		opt(publisher)
	}

	return publisher
}

// Start establishes the connection to the broker, declares the exchange, and sets up the publisher.
func (p *Publisher) Start(ctx context.Context) error {
	conn, err := amqp.ConnectWithTopicExchange(ctx, p.env, p.exchangeName)
	if err != nil {
		p.logger.Error("Failed to connect to topic exchange.", "err", err)

		return fmt.Errorf("connect to topic exchange: %w", err)
	}

	publisher, err := conn.NewPublisher(ctx, nil, nil)
	if err != nil {
		p.logger.Error("Failed to create publisher.", "err", err)
		err = fmt.Errorf("create publisher: %w", err)

		if closeErr := conn.Close(ctx); closeErr != nil {
			return errors.Join(err, fmt.Errorf("close AMQP connection: %w", closeErr))
		}

		return err
	}

	p.publisher = publisher

	p.mu.Lock()
	p.available = true
	p.mu.Unlock()

	return nil
}

// Publish sends a message with the given content to the configured exchange using the specified routing key.
func (p *Publisher) Publish(ctx context.Context, data []byte, key string) error {
	p.mu.RLock()

	if !p.available {
		p.mu.RUnlock()
		p.logger.Error("Publisher is not available")

		return ErrPublisherClosed
	}

	p.wg.Add(1)

	p.mu.RUnlock()
	defer p.wg.Done()

	msg, err := rmq.NewMessageWithAddress(data, &rmq.ExchangeAddress{
		Exchange: p.exchangeName,
		Key:      key,
	})
	if err != nil {
		p.logger.Error("Failed to publish message", "error", err)

		return fmt.Errorf("unable to publish message: %w", err)
	}

	res, err := p.publisher.Publish(ctx, msg)
	if err != nil {
		p.logger.Error("Failed to publish message", "error", err)

		return fmt.Errorf("unable to publish message: %w", err)
	}

	if err := CheckOutcome(res.Outcome); err != nil {
		p.logger.Error("Unexpected outcome", "error", err)

		return fmt.Errorf("unexpected outcome: %w", err)
	}

	p.logger.Debug("[x] Message sent", "content", data)

	return nil
}

// Close gracefully shuts down the publisher, waits for in-flight messages to complete, and closes the connections.
func (p *Publisher) Close(ctx context.Context) error {
	p.mu.Lock()
	p.available = false
	p.mu.Unlock()

	p.wg.Wait()
	p.logger.Debug("All in-flight publishes completed, proceeding to close.")

	ctx, cancel := context.WithTimeout(ctx, p.closeTimeout)
	defer cancel()

	if p.publisher != nil {
		err := p.publisher.Close(ctx)
		if err != nil {
			p.logger.Error("Failed to close publisher", "error", err)

			return fmt.Errorf("unable to close publisher: %w", err)
		}
	}

	if err := p.env.CloseConnections(ctx); err != nil {
		p.logger.Error("Failed to close connections", "error", err)

		return fmt.Errorf("unable to close connections: %w", err)
	}

	return nil
}

// WithPublisherLogger sets the logger for the Publisher.
func WithPublisherLogger(logger *slog.Logger) PublisherOption {
	return func(p *Publisher) {
		p.logger = logger
	}
}

// WithPublisherCloseTimeout sets the close timeout duration for the Publisher.
func WithPublisherCloseTimeout(timeout time.Duration) PublisherOption {
	return func(p *Publisher) {
		p.closeTimeout = timeout
	}
}
