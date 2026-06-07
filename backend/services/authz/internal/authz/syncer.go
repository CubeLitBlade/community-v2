package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq"
	v1 "github.com/cubelitblade/community-v2/backend/services/account/api/events/v1"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/constant"
)

// Syncer consumes events from RabbitMQ and synchronizes authorization tuples to OpenFGA.
type Syncer struct {
	subscriber *rabbitmq.QuorumSubscriber
	writer     *Writer
	logger     *slog.Logger

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewSyncer creates a Syncer with the given subscriber, writer, and logger.
func NewSyncer(subscriber *rabbitmq.QuorumSubscriber, writer *Writer, logger *slog.Logger) *Syncer {
	logger = logger.With(
		slog.String(constant.LogKeyService, constant.LogServiceAuthz),
		slog.String(constant.LogKeyComponent, constant.LogComponentSyncer),
	)

	return &Syncer{
		subscriber: subscriber,
		writer:     writer,
		logger:     logger,

		cancel: nil,
		wg:     sync.WaitGroup{},
	}
}

// Start begins consuming messages from the subscriber in a background goroutine.
func (s *Syncer) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	deliveries, err := s.subscriber.Start(ctx)
	if err != nil {
		s.cancel()
		s.logger.ErrorContext(ctx, "failed to start quorum subscriber", slog.Any(constant.LogKeyError, err))

		return fmt.Errorf("start quorum subscriber: %w", err)
	}

	s.wg.Add(1)
	go s.consumeLoop(ctx, deliveries)

	return nil
}

// Stop gracefully shuts down the Syncer, waiting for in-flight messages to complete or timing out.
func (s *Syncer) Stop(ctx context.Context) {
	if s.cancel != nil {
		s.cancel()
	}

	done := make(chan struct{})

	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.DebugContext(ctx, "syncer stopped gracefully")
	case <-ctx.Done():
		s.logger.WarnContext(ctx, "timed out waiting for syncer to stop, forcing exit")
	}
}

func (s *Syncer) consumeLoop(ctx context.Context, deliveries <-chan *rabbitmq.ReceivedMessage) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			s.logger.DebugContext(ctx, "shutting down, draining remaining deliveries")

			for delivery := range deliveries {
				s.processMessage(ctx, delivery)
			}

			return

		case delivery, ok := <-deliveries:
			if !ok {
				s.logger.DebugContext(ctx, "delivery channel closed, syncer exiting")
				return
			}

			s.processMessage(ctx, delivery)
		}
	}
}

func (s *Syncer) processMessage(ctx context.Context, delivery *rabbitmq.ReceivedMessage) {
	var event ce.Event
	if err := json.Unmarshal(delivery.Data, &event); err != nil {
		s.logger.ErrorContext(ctx, "poison message: failed to unmarshal cloud event",
			slog.Any(constant.LogKeyError, err),
		)

		if err := delivery.Accept(); err != nil {
			s.logger.ErrorContext(ctx, "failed to accept delivery with poison message",
				slog.Any(constant.LogKeyError, err),
			)
		}

		return
	}

	tuples, err := s.translateEvent(&event)
	if err != nil {
		s.logger.ErrorContext(ctx, "poison message: failed to translate event",
			slog.String(constant.LogKeyEventID, event.ID()),
			slog.Any(constant.LogKeyError, err),
		)

		if err := delivery.Accept(); err != nil {
			s.logger.ErrorContext(ctx, "failed to accept delivery ", slog.Any(constant.LogKeyError, err))
		}

		return
	}

	if len(tuples) == 0 {
		s.logger.DebugContext(ctx, "ignored event type", slog.String(constant.LogKeyEventType, event.Type()))

		if err := delivery.Accept(); err != nil {
			s.logger.ErrorContext(ctx, "failed to accept delivery with unexpected event type",
				slog.Any(constant.LogKeyError, err),
			)
		}

		return
	}

	if err := s.writer.WriteTuples(ctx, tuples); err != nil {
		s.logger.ErrorContext(ctx, "failed to write tuples, message will be retried",
			slog.String("event_id", event.ID()),
			slog.Any("error", err),
		)

		if err := delivery.Requeue(); err != nil {
			s.logger.ErrorContext(ctx, "failed to requeue delivery after writing tuple", slog.Any("error", err))
		}

		return
	}

	if err := delivery.Accept(); err != nil {
		s.logger.ErrorContext(ctx, "failed to accept delivery after successful write",
			slog.String("event_id", event.ID()),
			slog.Any("error", err),
		)
	}
}

func (s *Syncer) translateEvent(event *ce.Event) ([]Tuple, error) {
	switch event.Type() {
	case v1.EventTypeAccountCreated:
		return s.translateAccountCreated(event)
	default:
		return nil, nil
	}
}

func (s *Syncer) translateAccountCreated(event *ce.Event) ([]Tuple, error) {
	var payload v1.AccountCreatedEventPayload
	if err := event.DataAs(&payload); err != nil {
		return nil, fmt.Errorf("unmarshal account created data: %w", err)
	}

	tuples := []Tuple{
		{
			Subject:  "user:" + payload.AccountID,
			Relation: payload.Role,
			Object:   "system:community",
		},
	}

	return tuples, nil
}
