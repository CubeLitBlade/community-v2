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

var authDurationBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0}

type InstrumentedAuthenticator struct {
	inner         application.AuthenticatorUseCase
	tracer        trace.Tracer
	logger        *slog.Logger
	authCounter   metric.Int64Counter
	durationHist  metric.Float64Histogram
}

func NewInstrumentedAuthenticator(
	inner application.AuthenticatorUseCase,
	telemetry *Telemetry,
	logger *slog.Logger,
) application.AuthenticatorUseCase {
	meter := telemetry.Meter("account.application.authenticator")

	authCounter, _ := meter.Int64Counter("account.authenticate.total",
		metric.WithDescription("Total number of authentication attempts."),
		metric.WithUnit("{call}"),
	)

	durationHist, _ := meter.Float64Histogram("account.authenticate.duration_seconds",
		metric.WithDescription("Duration of authentication attempts in seconds."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(authDurationBuckets...),
	)

	return &InstrumentedAuthenticator{
		inner:        inner,
		tracer:       telemetry.Tracer("account.application.authenticator"),
		logger:       logger.With(slog.String("component", "authenticator_decorator")),
		authCounter:  authCounter,
		durationHist: durationHist,
	}
}

func (a *InstrumentedAuthenticator) Authenticate(ctx context.Context, username, password string) (*account.Account, error) {
	start := time.Now()

	ctx, span := a.tracer.Start(ctx, "Authenticator/Authenticate")
	defer span.End()

	acc, err := a.inner.Authenticate(ctx, username, password)

	outcome := "success"
	switch {
	case errors.Is(err, account.ErrInvalidCredentials):
		outcome = "invalid_credentials"
	case errors.Is(err, account.ErrAccountNotFound):
		outcome = "not_found"
	case err != nil:
		outcome = "error"
		span.SetStatus(codes.Error, "authentication failed")
		span.RecordError(err)
	}

	a.authCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("outcome", outcome),
	))
	a.durationHist.Record(ctx, time.Since(start).Seconds())

	if err != nil {
		return nil, fmt.Errorf("instrumented authenticate: %w", err)
	}

	return acc, nil
}
