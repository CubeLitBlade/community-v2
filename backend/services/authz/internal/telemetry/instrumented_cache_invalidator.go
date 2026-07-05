package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

type InstrumentedCacheInvalidator struct {
	inner    port.CacheInvalidator
	total    metric.Int64Counter
	duration metric.Float64Histogram
}

func NewInstrumentedCacheInvalidator(inner port.CacheInvalidator, telemetry *Telemetry) port.CacheInvalidator {
	meter := telemetry.Meter("authz.adapter.cache_invalidator")

	total, _ := meter.Int64Counter("authz.cache.invalidation_total",
		metric.WithDescription("Total number of cache invalidations."),
		metric.WithUnit("{invalidation}"),
	)

	duration, _ := meter.Float64Histogram("authz.cache.invalidation.duration_seconds",
		metric.WithDescription("Duration of cache invalidations in seconds."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(cacheDurationBuckets...),
	)

	return &InstrumentedCacheInvalidator{
		inner:    inner,
		total:    total,
		duration: duration,
	}
}

func (i *InstrumentedCacheInvalidator) Invalidate(ctx context.Context, keys []string) error {
	start := time.Now()
	err := i.inner.Invalidate(ctx, keys)

	i.total.Add(ctx, 1)
	i.duration.Record(ctx, time.Since(start).Seconds())

	if err != nil {
		return fmt.Errorf("instrumented cache invalidate: %w", err)
	}

	return nil
}
