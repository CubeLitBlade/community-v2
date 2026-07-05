package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/config"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/telemetry"
)

func NewOpenTelemetryTraceProvider(
	cfg config.OTELConfig, lc fx.Lifecycle, logger *slog.Logger,
) (trace.TracerProvider, error) {
	if !cfg.Enable {
		logger.Info("otel disabled")

		tp := noop.NewTracerProvider()
		otel.SetTracerProvider(tp)

		return tp, nil
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.InfoContext(ctx, "shutting down TracerProvider")
			return provider.Shutdown(ctx)
		},
	})

	logger.Info("otel tracer provider initialized", slog.String("endpoint", cfg.Endpoint))

	return provider, nil
}

const metricExportInterval = 15 * time.Second // TODO: make configurable via OTEL config

func NewOpenTelemetryMeterProvider(
	cfg config.OTELConfig, lc fx.Lifecycle, logger *slog.Logger,
) (metric.MeterProvider, error) {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	exporter, err := otlpmetricgrpc.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp metric exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(metricExportInterval)),
		),
	)

	otel.SetMeterProvider(provider)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.InfoContext(ctx, "shutting down MeterProvider")
			return provider.Shutdown(ctx)
		},
	})

	logger.Info("otel meter provider initialized", slog.String("endpoint", cfg.Endpoint))

	return provider, nil
}

func OpenTelemetryModule() fx.Option {
	return fx.Options(
		fx.Provide(NewOpenTelemetryTraceProvider),
		fx.Provide(NewOpenTelemetryMeterProvider),
		fx.Provide(func(mp metric.MeterProvider, tp trace.TracerProvider) *telemetry.Telemetry {
			return &telemetry.Telemetry{
				MeterProvider:  mp,
				TracerProvider: tp,
			}
		}),
	)
}
