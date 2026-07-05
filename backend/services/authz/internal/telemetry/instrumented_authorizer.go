package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/application"
)

var authzDurationBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0}

type InstrumentedAuthorizer struct {
	inner            application.AuthChecker
	tracer           trace.Tracer
	logger           *slog.Logger
	decisionsCounter metric.Int64Counter
	durationHist     metric.Float64Histogram
}

func NewInstrumentedAuthorizer(
	inner application.AuthChecker,
	telemetry *Telemetry,
	logger *slog.Logger,
) application.AuthChecker {
	meter := telemetry.Meter("authz.application.authorizer")

	decisionsCounter, _ := meter.Int64Counter("authz.check.decisions_total",
		metric.WithDescription("Total number of authorization decisions."),
		metric.WithUnit("{decision}"),
	)

	durationHist, _ := meter.Float64Histogram("authz.check.duration_seconds",
		metric.WithDescription("Duration of authorization checks in seconds."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(authzDurationBuckets...),
	)

	return &InstrumentedAuthorizer{
		inner:            inner,
		tracer:           telemetry.Tracer("authz.application.authorizer"),
		logger:           logger.With(slog.String("component", "authorizer_decorator")),
		decisionsCounter: decisionsCounter,
		durationHist:     durationHist,
	}
}

func (a *InstrumentedAuthorizer) Authorize(ctx context.Context, user, relation, object string) (bool, error) {
	start := time.Now()

	ctx, span := a.tracer.Start(ctx, "Authorizer/Authorize")
	defer span.End()

	span.SetAttributes(
		attribute.String("authz.user", user),
		attribute.String("authz.relation", relation),
		attribute.String("authz.object", object),
	)

	allowed, err := a.inner.Authorize(ctx, user, relation, object)

	var decision string

	switch {
	case err != nil:
		decision = "error"

		span.SetStatus(codes.Error, "failed to authorize")
		span.RecordError(err)
	case allowed:
		decision = "allow"
	default:
		decision = "deny"
	}

	span.SetAttributes(attribute.Bool("authz.allowed", allowed))

	a.decisionsCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("decision", decision),
	))
	a.durationHist.Record(ctx, time.Since(start).Seconds())

	if err != nil {
		return allowed, fmt.Errorf("instrumented authorize: %w", err)
	}

	return allowed, nil
}
