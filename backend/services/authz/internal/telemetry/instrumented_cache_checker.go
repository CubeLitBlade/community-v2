package telemetry

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/metric"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

type InstrumentedCacheChecker struct {
	inner  port.CacheChecker
	hits   metric.Int64Counter
	misses metric.Int64Counter
	errs   metric.Int64Counter
}

func NewInstrumentedCacheChecker(inner port.CacheChecker, telemetry *Telemetry) port.CacheChecker {
	meter := telemetry.Meter("authz.adapter.cache_checker")

	hits, _ := meter.Int64Counter("authz.cache.hits_total",
		metric.WithDescription("Total number of cache hits."),
		metric.WithUnit("{hit}"),
	)

	misses, _ := meter.Int64Counter("authz.cache.misses_total",
		metric.WithDescription("Total number of cache misses."),
		metric.WithUnit("{miss}"),
	)

	errs, _ := meter.Int64Counter("authz.cache.errors_total",
		metric.WithDescription("Total number of cache check errors."),
		metric.WithUnit("{error}"),
	)

	return &InstrumentedCacheChecker{
		inner:  inner,
		hits:   hits,
		misses: misses,
		errs:   errs,
	}
}

func (c *InstrumentedCacheChecker) Check(ctx context.Context, user, relation, object string) (bool, error) {
	allowed, err := c.inner.Check(ctx, user, relation, object)

	switch {
	case err == nil:
		c.hits.Add(ctx, 1)
	case errors.Is(err, port.ErrCacheMiss):
		c.misses.Add(ctx, 1)
	default:
		c.errs.Add(ctx, 1)
	}

	if err != nil {
		return allowed, fmt.Errorf("instrumented cache check: %w", err)
	}

	return allowed, nil
}
