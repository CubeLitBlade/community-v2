package application

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

type TupleWriter struct {
	writer port.TupleWriter
	logger *slog.Logger
}

func NewTupleWriter(writer port.TupleWriter, logger *slog.Logger) *TupleWriter {
	return &TupleWriter{
		writer: writer,
		logger: logger.With(slog.String("component", "tuple_writer")),
	}
}

func (w *TupleWriter) Write(ctx context.Context, tuples []domain.Tuple) error {
	if len(tuples) == 0 {
		return nil
	}

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int("authz.tuples_count", len(tuples)),
		attribute.String("authz.first_tuple_user", tuples[0].User),
	)

	if err := w.writer.Write(ctx, tuples); err != nil {
		w.logger.ErrorContext(ctx, "failed to land tuples", slog.Any("error", err))
		span.SetStatus(codes.Error, "failed to write tuples")
		span.RecordError(err)

		return fmt.Errorf("write tuples: %w", err)
	}

	w.logger.InfoContext(ctx, "tuples settled", slog.Int("count", len(tuples)))

	return nil
}
