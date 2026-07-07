package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/application"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
)

var profileDurationBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0}

type InstrumentedProfileFinder struct {
	inner        application.ProfileFinderUseCase
	tracer       trace.Tracer
	logger       *slog.Logger
	findCounter  metric.Int64Counter
	durationHist metric.Float64Histogram
}

func NewInstrumentedProfileFinder(
	inner application.ProfileFinderUseCase,
	telemetry *Telemetry,
	logger *slog.Logger,
) application.ProfileFinderUseCase {
	meter := telemetry.Meter("account.application.profile_finder")

	findCounter, _ := meter.Int64Counter("account.profile_find.total",
		metric.WithDescription("Total number of profile lookups."),
		metric.WithUnit("{call}"),
	)

	durationHist, _ := meter.Float64Histogram("account.profile_find.duration_seconds",
		metric.WithDescription("Duration of profile lookups in seconds."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(profileDurationBuckets...),
	)

	return &InstrumentedProfileFinder{
		inner:        inner,
		tracer:       telemetry.Tracer("account.application.profile_finder"),
		logger:       logger.With(slog.String("component", "profile_finder_decorator")),
		findCounter:  findCounter,
		durationHist: durationHist,
	}
}

func (f *InstrumentedProfileFinder) Find(ctx context.Context, id int64) (*application.Profile, error) {
	start := time.Now()

	ctx, span := f.tracer.Start(ctx, "ProfileFinder/Find")
	defer span.End()

	span.SetAttributes(attribute.Int64("account.id", id))

	profile, err := f.inner.Find(ctx, id)

	outcome := "success"
	switch {
	case errors.Is(err, account.ErrAccountNotFound):
		outcome = "not_found"
	case err != nil:
		outcome = "error"
		span.SetStatus(codes.Error, "profile lookup failed")
		span.RecordError(err)
	}

	f.findCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("outcome", outcome),
	))
	f.durationHist.Record(ctx, time.Since(start).Seconds())

	if err != nil {
		return nil, fmt.Errorf("instrumented find profile: %w", err)
	}

	return profile, nil
}
