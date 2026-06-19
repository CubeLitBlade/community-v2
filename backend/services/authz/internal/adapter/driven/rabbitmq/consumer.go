package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/cloudevents/sdk-go/v2/event"

	rmq "github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/application"
)

type Consumer struct {
	subscriber *rmq.QuorumSubscriber
	handler    *application.EventSyncer
	logger     *slog.Logger

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewConsumer(
	subscriber *rmq.QuorumSubscriber, handler *application.EventSyncer, logger *slog.Logger,
) *Consumer {
	return &Consumer{
		subscriber: subscriber,
		handler:    handler,
		logger:     logger.With(slog.String("component", "rabbitmq_consumer")),

		cancel: nil,
		wg:     sync.WaitGroup{},
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	c.logger.InfoContext(ctx, "rabbitmq consumer is ready")

	ctx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	c.cancel = cancel

	deliveries, err := c.subscriber.Start(ctx)
	if err != nil {
		c.cancel()
		c.logger.ErrorContext(ctx, "failed to start quorum subscriber", slog.Any("error", err))

		return fmt.Errorf("start quorum subscriber: %w", err)
	}

	c.wg.Add(1)
	go c.consume(ctx, deliveries)

	return nil
}

func (c *Consumer) Stop(ctx context.Context) {
	if c.cancel != nil {
		c.cancel()
	}

	done := make(chan struct{})

	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.logger.InfoContext(ctx, "rabbitmq consumer stopped gracefully")
	case <-ctx.Done():
		c.logger.WarnContext(ctx, "timed out waiting for consumer to stop, forcing exit")
	}
}

func (c *Consumer) consume(ctx context.Context, deliveries <-chan *rmq.ReceivedMessage) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			c.logger.DebugContext(ctx, "shutting down, draining remaining deliveries")

			c.drainDeliveries(ctx, deliveries)

			return

		case delivery, ok := <-deliveries:
			if !ok {
				c.logger.WarnContext(ctx, "delivery channel closed, consumer exiting")
				return
			}

			c.process(ctx, delivery)
		}
	}
}

func (c *Consumer) process(ctx context.Context, delivery *rmq.ReceivedMessage) {
	var evt event.Event
	if err := sonic.Unmarshal(delivery.Data, &evt); err != nil {
		c.logger.ErrorContext(ctx, "poison message: failed to unmarshal cloud event", slog.Any("error", err))

		if err := delivery.Accept(); err != nil {
			c.logger.ErrorContext(ctx, "failed to accept delivery", slog.Any("error", err))
		}

		return
	}

	translated, err := translate(evt)
	if err != nil {
		c.logger.ErrorContext(ctx, "poison message: failed to translate event",
			slog.String("event_id", evt.ID()),
			slog.Any("error", err),
		)

		c.accept(ctx, delivery, evt.ID())

		return
	}

	if err := c.handler.Sync(ctx, translated); err != nil {
		c.logger.ErrorContext(ctx, "failed to handle event, requeuing message",
			slog.String("event_id", evt.ID()),
			slog.Any("error", err),
		)

		c.requeue(ctx, delivery, evt.ID())

		return
	}

	c.accept(ctx, delivery, evt.ID())
}

func (c *Consumer) drainDeliveries(ctx context.Context, deliveries <-chan *rmq.ReceivedMessage) {
	for d := range deliveries {
		c.process(ctx, d)
	}
}

func (c *Consumer) accept(ctx context.Context, delivery *rmq.ReceivedMessage, eventID string) {
	if err := delivery.Accept(); err != nil {
		c.logger.ErrorContext(ctx, "fatal: failed to accept delivery",
			slog.String("event_id", eventID),
			slog.Any("error", err),
		)
	}
}

func (c *Consumer) requeue(ctx context.Context, delivery *rmq.ReceivedMessage, eventID string) {
	if err := delivery.Requeue(); err != nil {
		c.logger.ErrorContext(ctx, "fatal: failed to requeue delivery",
			slog.String("event_id", eventID),
			slog.Any("error", err),
		)
	}
}
