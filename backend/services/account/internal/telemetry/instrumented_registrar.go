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

	"github.com/cubelitblade/community-v2/backend/services/account/internal/application"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
)

var registerDurationBuckets = []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0}

type InstrumentedRegistrar struct {
	inner            application.RegistrarUseCase
	tracer           trace.Tracer
	logger           *slog.Logger
	registerCounter  metric.Int64Counter
	durationHist     metric.Float64Histogram
}

func NewInstrumentedRegistrar(
	inner application.RegistrarUseCase,
	telemetry *Telemetry,
	logger *slog.Logger,
) application.RegistrarUseCase {
	meter := telemetry.Meter("account.application.registrar")

	registerCounter, _ := meter.Int64Counter("account.register.total",
		metric.WithDescription("Total number of account registrations."),
		metric.WithUnit("{call}"),
	)

	durationHist, _ := meter.Float64Histogram("account.register.duration_seconds",
		metric.WithDescription("Duration of account registrations in seconds."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(registerDurationBuckets...),
	)

	return &InstrumentedRegistrar{
		inner:           inner,
		tracer:          telemetry.Tracer("account.application.registrar"),
		logger:          logger.With(slog.String("component", "registrar_decorator")),
		registerCounter: registerCounter,
		durationHist:    durationHist,
	}
}

func (r *InstrumentedRegistrar) Register(ctx context.Context, username string, password string) (*account.Account, error) {
	start := time.Now()

	ctx, span := r.tracer.Start(ctx, "Registrar/Register")
	defer span.End()

	acc, err := r.inner.Register(ctx, username, password)

	outcome := "success"
	if err != nil {
		outcome = "error"
		span.SetStatus(codes.Error, "registration failed")
		span.RecordError(err)
	}

	r.registerCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("outcome", outcome),
	))
	r.durationHist.Record(ctx, time.Since(start).Seconds())

	if err != nil {
		return nil, fmt.Errorf("instrumented register: %w", err)
	}

	return acc, nil
}
