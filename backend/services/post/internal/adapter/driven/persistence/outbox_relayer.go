package persistence

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	outboxevent "github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence/ent"
	entoutbox "github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence/ent/outbox"
)

const (
	defaultScanLimit   = 10
	defaultMinInterval = 1 * time.Second
	defaultMaxInterval = 4 * time.Second
)

// Publisher publishes event data to an external message broker.
type Publisher interface {
	Publish(ctx context.Context, data []byte, key string) error
}

// OutboxRelay polls unpublished outbox entries, publishes them, and acks them.
type OutboxRelay struct {
	ids         idgen.Generator
	client      *ent.Client
	publisher   Publisher
	minInterval time.Duration
	maxInterval time.Duration
	logger      *slog.Logger

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// NewOutboxRelay creates a new OutboxRelay instance.
func NewOutboxRelay(ids idgen.Generator, client *ent.Client, publisher Publisher, logger *slog.Logger) *OutboxRelay {
	return &OutboxRelay{
		ids:         ids,
		client:      client,
		publisher:   publisher,
		minInterval: defaultMinInterval,
		maxInterval: defaultMaxInterval,
		logger:      logger.With(slog.String("module", "outbox_relay")),
	}
}

// Start begins the scan and publish loops in background goroutines.
func (r *OutboxRelay) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	r.cancel = cancel

	r.wg.Add(1)
	go r.runLoop(ctx)
}

// Stop gracefully shuts down the relay.
func (r *OutboxRelay) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
	r.wg.Wait()
}

func (r *OutboxRelay) runLoop(ctx context.Context) {
	defer r.wg.Done()

	currentInterval := r.minInterval
	timer := time.NewTimer(currentInterval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			count, err := r.publishPending(ctx)
			if err != nil {
				currentInterval = min(currentInterval*2, r.maxInterval)
				r.logger.ErrorContext(ctx, "scan failed", slog.Any("error", err))
			} else if count > 0 {
				currentInterval = r.minInterval
			} else {
				currentInterval = min(currentInterval*2, r.maxInterval)
			}
			timer.Reset(currentInterval)
		}
	}
}

func (r *OutboxRelay) publishPending(ctx context.Context) (int, error) {
	entries, err := r.client.Outbox.Query().
		Where(entoutbox.PublishedAtIsNil()).
		Order(ent.Asc(entoutbox.FieldCreatedAt)).
		Limit(defaultScanLimit).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("query outbox: %w", err)
	}

	for _, entry := range entries {
		if err := r.publishEntry(ctx, entry); err != nil {
			return len(entries), fmt.Errorf("publish entry %d: %w", entry.ID, err)
		}
	}

	return len(entries), nil
}

func (r *OutboxRelay) publishEntry(ctx context.Context, entry *ent.Outbox) error {
	eventID, err := r.ids.NextID()
	if err != nil {
		return fmt.Errorf("generate event id: %w", err)
	}

	payload := entry.Payload

	event, err := outboxevent.NewEvent(
		entry.AggregateType, entry.EventType, payload, eventID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("create cloud event: %w", err)
	}

	if err := r.publisher.Publish(ctx, event, entry.Topic); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	now := time.Now()
	if _, err := r.client.Outbox.Update().
		Where(entoutbox.IDEQ(entry.ID)).
		SetPublishedAt(now).
		Save(ctx); err != nil {
		return fmt.Errorf("ack entry %d: %w", entry.ID, err)
	}

	r.logger.InfoContext(ctx, "outbox entry published",
		slog.Int64("id", entry.ID),
		slog.String("topic", entry.Topic),
		slog.String("event_type", entry.EventType),
	)

	return nil
}
