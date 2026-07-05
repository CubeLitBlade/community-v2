package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

type InstrumentedCacheWriter struct {
	inner    port.CacheWriter
	total    metric.Int64Counter
	duration metric.Float64Histogram
}

func NewInstrumentedCacheWriter(inner port.CacheWriter, telemetry *Telemetry) port.CacheWriter {
	meter := telemetry.Meter("authz.adapter.cache_writer")

	total, _ := meter.Int64Counter("authz.cache.write_total",
		metric.WithDescription("Total number of cache writes."),
		metric.WithUnit("{write}"),
	)

	duration, _ := meter.Float64Histogram("authz.cache.write.duration_seconds",
		metric.WithDescription("Duration of cache writes in seconds."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(cacheDurationBuckets...),
	)

	return &InstrumentedCacheWriter{
		inner:    inner,
		total:    total,
		duration: duration,
	}
}

func (w *InstrumentedCacheWriter) Write(
	ctx context.Context, user, relation, object string, allowed bool, expiration time.Duration,
) error {
	start := time.Now()
	err := w.inner.Write(ctx, user, relation, object, allowed, expiration)

	w.total.Add(ctx, 1)
	w.duration.Record(ctx, time.Since(start).Seconds())

	if err != nil {
		return fmt.Errorf("instrumented cache write: %w", err)
	}

	return nil
}
