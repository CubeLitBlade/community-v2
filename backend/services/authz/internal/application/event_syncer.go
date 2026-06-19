package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain"
)

type EventSyncer struct {
	writer *TupleWriter
	logger *slog.Logger
}

func NewEventSyncer(write *TupleWriter, logger *slog.Logger) *EventSyncer {
	return &EventSyncer{
		writer: write,
		logger: logger.With(slog.String("component", "event_syncer")),
	}
}

func (s *EventSyncer) Sync(ctx context.Context, evt domain.Event) error {
	switch e := evt.(type) {
	case *domain.AccountCreated:
		return s.syncAccountCreatedEvent(ctx, e)
	default:
		s.logger.WarnContext(ctx, "unfamiliar event type, letting it slide",
			slog.String("type", evt.Type()))

		return nil
	}
}

func (s *EventSyncer) syncAccountCreatedEvent(ctx context.Context, evt *domain.AccountCreated) error {
	if err := s.writer.Write(ctx, evt.Tuples); err != nil {
		return fmt.Errorf("write tuples: %w", err)
	}

	return nil
}
