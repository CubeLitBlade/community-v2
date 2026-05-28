package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/events/internal/amqp"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

// PublisherConfig holds the configuration options for initializing a Publisher.
type PublisherConfig struct {
	BrokerURL    string
	Logger       *slog.Logger
	ExchangeName string
	CloseTimeout time.Duration
}

// Publisher manages the publishing of messages to a specific RabbitMQ exchange.
type Publisher struct {
	cfg    *PublisherConfig
	logger *slog.Logger

	conn      *rmq.AmqpConnection
	env       *rmq.Environment
	publisher *rmq.Publisher

	mu        sync.RWMutex
	available bool
	wg        sync.WaitGroup
}

// NewPublisher creates and returns a new instance of Publisher.
func NewPublisher(cfg *PublisherConfig) *Publisher {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	if cfg.CloseTimeout == 0 {
		cfg.CloseTimeout = DefaultCloseTimeout
	}

	return &Publisher{
		cfg:       cfg,
		logger:    cfg.Logger,
		conn:      nil,
		env:       rmq.NewEnvironment(cfg.BrokerURL, nil),
		publisher: nil,
		mu:        sync.RWMutex{},
		available: false,
		wg:        sync.WaitGroup{},
	}
}

// Start establishes the connection to the broker, declares the exchange, and sets up the publisher.
func (p *Publisher) Start(ctx context.Context) error {
	conn, err := amqp.Connect(ctx, p.env)
	if err != nil {
		if err := p.env.CloseConnections(ctx); err != nil {
			p.logger.Error("Failed to close connections", "error", err)
		}
		p.logger.Error("Failed to connect to RabbitMQ.")
		return fmt.Errorf("connect AMQP connection: %w", err)
	}

	if err := amqp.DeclareTopicExchange(ctx, conn, p.cfg.ExchangeName); err != nil {
		p.closeConn(ctx)
		p.logger.Error("Failed to declare topic exchange.")
		return fmt.Errorf("declare topic exchange: %w", err)
	}

	publisher, err := conn.NewPublisher(ctx, nil, nil)
	if err != nil {
		p.closeConn(ctx)
		p.logger.Error("Failed to create publisher", "error", err)
		return fmt.Errorf("unable to create publisher: %w", err)
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
		Exchange: p.cfg.ExchangeName,
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
func (p *Publisher) Close() error {
	p.mu.Lock()
	p.available = false
	p.mu.Unlock()

	p.wg.Wait()
	p.logger.Debug("All in-flight publishes completed, proceeding to close.")

	ctx, cancel := context.WithTimeout(context.Background(), p.cfg.CloseTimeout)
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

func (p *Publisher) closeConn(ctx context.Context) {
	if err := p.env.CloseConnections(ctx); err != nil {
		p.logger.Error("Failed to close connections", "error", err)
	}
}
