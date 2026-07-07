package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
)

const (
	defaultChannelBufferSize = 32
	defaultScanLimit         = 10
)

// RelayOption applies a configuration to a Relay.
type RelayOption func(*Relay)

// Scanner scans for pending outbox entries in the database.
type Scanner interface {
	Scan(ctx context.Context, db *gorm.DB, count int) ([]*Entry, error)
}

// Recorder records the successful publication of an outbox entry.
type Recorder interface {
	Ack(ctx context.Context, db *gorm.DB, id int64, now time.Time) error
}

// Publisher publishes event data to an external message broker.
type Publisher interface {
	Publish(ctx context.Context, data []byte, key string) error
}

// Relay orchestrates the transactional outbox pattern,
// scanning for unpublished entries and publishing them asynchronously.
type Relay struct {
	ids       idgen.Generator
	scanner   Scanner
	recorder  Recorder
	publisher Publisher
	channel   chan *Entry
	db        *gorm.DB

	minInterval time.Duration
	maxInterval time.Duration
	clock       func() time.Time
	logger      *slog.Logger

	wg     sync.WaitGroup
	cancel context.CancelFunc
	once   sync.Once
}

// NewRelay creates a new Relay instance with the specified dependencies and options.
func NewRelay(ids idgen.Generator, scanner Scanner, recorder Recorder,
	publisher Publisher, db *gorm.DB, logger *slog.Logger, opts ...RelayOption,
) *Relay {
	logger = logger.With(slog.String("module", "outbox"))

	relay := &Relay{
		ids:         ids,
		scanner:     scanner,
		recorder:    recorder,
		publisher:   publisher,
		channel:     make(chan *Entry, defaultChannelBufferSize),
		db:          db,
		minInterval: 1 * time.Second,
		maxInterval: 4 * time.Second,
		clock:       time.Now,
		logger:      logger,
		wg:          sync.WaitGroup{},
		cancel:      nil,
		once:        sync.Once{},
	}

	for _, opt := range opts {
		opt(relay)
	}

	return relay
}

// Start begins the scan and publish loops in separate goroutines.
func (r *Relay) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	r.cancel = cancel

	go r.scanLoop(ctx, defaultScanLimit)
	go r.workLoop(ctx)
}

// Close gracefully shuts down the relay, waiting for ongoing operations to complete.
func (r *Relay) Close() error {
	r.once.Do(func() {
		r.cancel()
		r.wg.Wait()
		close(r.channel)
	})

	return nil
}

func (r *Relay) scanLoop(ctx context.Context, count int) {
	currentInterval := r.minInterval

	timer := time.NewTimer(currentInterval)
	defer timer.Stop()

	r.wg.Add(1)
	defer r.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			entries, err := r.scanner.Scan(ctx, r.db, count)

			switch {
			case err != nil:
				currentInterval = min(currentInterval*2, r.maxInterval)
				r.logger.ErrorContext(ctx, "scan failed", slog.Any("error", err))
			case len(entries) > 0:
				currentInterval = r.minInterval
				r.logger.InfoContext(ctx, "relaying events", slog.Int("count", len(entries)))

				for _, entry := range entries {
					select {
					case r.channel <- entry:
					case <-ctx.Done():
						return
					}
				}
			default:
				currentInterval = min(currentInterval*2, r.maxInterval)
			}

			timer.Reset(currentInterval)
		}
	}
}

func (r *Relay) workLoop(ctx context.Context) {
	r.wg.Add(1)
	defer r.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case entry := <-r.channel:
			if err := r.publish(ctx, entry); err != nil {
				r.logger.ErrorContext(ctx, "publish failed", slog.Any("error", err))
				continue
			}

			err := r.recorder.Ack(ctx, r.db, entry.ID, r.clock())
			if err != nil {
				r.logger.ErrorContext(ctx, "ack failed",
					slog.Int64("id", entry.ID),
					slog.Any("error", err))
				continue
			}
		}
	}
}

func (r *Relay) publish(ctx context.Context, entry *Entry) error {
	id, err := r.ids.NextID()
	if err != nil {
		r.logger.WarnContext(ctx, "id generation failed", slog.Any("error", err))
		return fmt.Errorf("generate event ID: %w", err)
	}

	event, err := NewEvent(
		entry.AggregateType, entry.EventType, entry.Payload, id, r.clock(),
	)
	if err != nil {
		return fmt.Errorf("create event: %w", err)
	}

	if err := r.publisher.Publish(ctx, event, entry.Topic); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}

	return nil
}

// WithClock sets a custom clock function for the relay.
func WithClock(clock func() time.Time) RelayOption {
	return func(r *Relay) {
		r.clock = clock
	}
}

// WithMinInterval sets the minimum scan interval.
func WithMinInterval(interval time.Duration) RelayOption {
	return func(r *Relay) {
		r.minInterval = interval
	}
}

// WithMaxInterval sets the maximum scan interval for exponential backoff.
func WithMaxInterval(interval time.Duration) RelayOption {
	return func(r *Relay) {
		r.maxInterval = interval
	}
}
