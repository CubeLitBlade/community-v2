package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

const defaultPublisherCloseTimeout = 30 * time.Second

var (
	// ErrMessageStateRejected indicates that the published message was rejected by the broker.
	ErrMessageStateRejected = errors.New("message was rejected")

	// ErrMessageStateReleased indicates that the published message was released by the broker.
	ErrMessageStateReleased = errors.New("message was released")

	// ErrMessageStateModified indicates that the published message was modified by the broker.
	ErrMessageStateModified = errors.New("message was modified")

	// ErrMessageStateUnknown indicates that the broker returned an unknown message outcome state.
	ErrMessageStateUnknown = errors.New("message state unknown")

	// ErrPublisherClosed indicates that a publish operation was attempted while the publisher is closed.
	ErrPublisherClosed = errors.New("publisher is closed")
)

// ExchangePublisherConfig holds the configuration options for initializing an ExchangePublisher.
type ExchangePublisherConfig struct {
	BrokerURL    string
	Logger       *slog.Logger
	ExchangeName string
	CloseTimeout time.Duration
}

// ExchangePublisher manages the publishing of messages to a specific RabbitMQ exchange.
type ExchangePublisher struct {
	cfg    *ExchangePublisherConfig
	logger *slog.Logger

	conn      *rmq.AmqpConnection
	env       *rmq.Environment
	publisher *rmq.Publisher

	mu        sync.RWMutex
	available bool
	wg        sync.WaitGroup
}

// NewExchangePublisher creates and returns a new instance of ExchangePublisher.
func NewExchangePublisher(cfg *ExchangePublisherConfig) *ExchangePublisher {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	if cfg.CloseTimeout == 0 {
		cfg.CloseTimeout = defaultPublisherCloseTimeout
	}

	return &ExchangePublisher{
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

// Init establishes the connection to the broker, declares the exchange, and sets up the publisher.
func (p *ExchangePublisher) Init(ctx context.Context) error {
	conn, err := p.env.NewConnection(ctx)
	if err != nil {
		if err := p.env.CloseConnections(ctx); err != nil {
			p.logger.Error("Failed to close connections", "error", err)
		}
		p.logger.Error("Failed to connect to RabbitMQ", "error", err)
		return fmt.Errorf("unable to connect to broker: %w", err)
	}
	p.conn = conn

	_, err = conn.Management().DeclareExchange(ctx, &rmq.DirectExchangeSpecification{
		Name: p.cfg.ExchangeName,
	})
	if err != nil {
		p.closeConn(ctx)
		p.logger.Error("Failed to declare exchange", "error", err)
		return fmt.Errorf("unable to declare exchange: %w", err)
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
func (p *ExchangePublisher) Publish(ctx context.Context, content string, key string) error {
	p.mu.RLock()
	if !p.available {
		p.mu.RUnlock()
		p.logger.Error("Publisher is not available")
		return ErrPublisherClosed
	}
	p.wg.Add(1)
	p.mu.RUnlock()
	defer p.wg.Done()

	msg, err := rmq.NewMessageWithAddress([]byte(content), &rmq.ExchangeAddress{
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

	switch res.Outcome.(type) {
	case *rmq.StateAccepted:
		p.logger.Debug("Published state accepted", "outcome", res.Outcome)
	case *rmq.StateRejected:
		p.logger.Debug("Published state rejected", "outcome", res.Outcome)
		return ErrMessageStateRejected
	case *rmq.StateReleased:
		p.logger.Debug("Published state released", "outcome", res.Outcome)
		return ErrMessageStateReleased
	case *rmq.StateModified:
		p.logger.Debug("Published state modified", "outcome", res.Outcome)
		return ErrMessageStateModified
	default:
		p.logger.Error("Published state unknown", "outcome", res.Outcome)
		return ErrMessageStateUnknown
	}

	p.logger.Debug("[x] Message sent", "content", content)
	return nil
}

// Close gracefully shuts down the publisher, waits for in-flight messages to complete, and closes the connections.
func (p *ExchangePublisher) Close() error {
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

func (p *ExchangePublisher) closeConn(ctx context.Context) {
	if err := p.env.CloseConnections(ctx); err != nil {
		p.logger.Error("Failed to close connections", "error", err)
	}
}
