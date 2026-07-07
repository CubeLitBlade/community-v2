package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain/port"
)

var outboxDurationBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0}

type InstrumentedEventRecorder struct {
	inner        port.EventRecorder
	saveCounter  metric.Int64Counter
	durationHist metric.Float64Histogram
}

func NewInstrumentedEventRecorder(inner port.EventRecorder, telemetry *Telemetry) port.EventRecorder {
	meter := telemetry.Meter("account.application.event_recorder")

	saveCounter, _ := meter.Int64Counter("account.outbox.save.total",
		metric.WithDescription("Total number of outbox events persisted."),
		metric.WithUnit("{event}"),
	)

	durationHist, _ := meter.Float64Histogram("account.outbox.save.duration_seconds",
		metric.WithDescription("Duration of outbox event persistence in seconds."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(outboxDurationBuckets...),
	)

	return &InstrumentedEventRecorder{
		inner:        inner,
		saveCounter:  saveCounter,
		durationHist: durationHist,
	}
}

func (r *InstrumentedEventRecorder) Record(ctx context.Context, entry *outbox.Entry) error {
	start := time.Now()

	err := r.inner.Record(ctx, entry)

	outcome := "success"
	if err != nil {
		outcome = "error"
	}

	r.saveCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("outcome", outcome),
		attribute.String("topic", entry.Topic),
	))
	r.durationHist.Record(ctx, time.Since(start).Seconds())

	if err != nil {
		return fmt.Errorf("instrumented event record: %w", err)
	}

	return nil
}
