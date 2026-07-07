package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
)

var relayDurationBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0}

type InstrumentedPublisher struct {
	inner          outbox.Publisher
	publishCounter metric.Int64Counter
	durationHist   metric.Float64Histogram
}

func NewInstrumentedPublisher(inner outbox.Publisher, telemetry *Telemetry) outbox.Publisher {
	meter := telemetry.Meter("account.adapter.outbox_publisher")

	publishCounter, _ := meter.Int64Counter("account.outbox.relay.publish_total",
		metric.WithDescription("Total number of outbox relay publish attempts."),
		metric.WithUnit("{message}"),
	)

	durationHist, _ := meter.Float64Histogram("account.outbox.relay.publish.duration_seconds",
		metric.WithDescription("Duration of outbox relay publishes in seconds."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(relayDurationBuckets...),
	)

	return &InstrumentedPublisher{
		inner:          inner,
		publishCounter: publishCounter,
		durationHist:   durationHist,
	}
}

func (p *InstrumentedPublisher) Publish(ctx context.Context, data []byte, key string) error {
	start := time.Now()

	err := p.inner.Publish(ctx, data, key)

	outcome := "success"
	if err != nil {
		outcome = "error"
	}

	p.publishCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("outcome", outcome),
		attribute.String("topic", key),
	))
	p.durationHist.Record(ctx, time.Since(start).Seconds())

	if err != nil {
		return fmt.Errorf("instrumented publish: %w", err)
	}

	return nil
}
