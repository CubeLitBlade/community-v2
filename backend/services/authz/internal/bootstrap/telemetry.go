package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/config"
)

func NewOpenTelemetryTraceProvider(
	cfg config.OTELConfig, lc fx.Lifecycle, logger *slog.Logger,
) (*sdktrace.TracerProvider, error) {
	if !cfg.Enable {
		logger.Info("otel disabled")

		tp := sdktrace.NewTracerProvider()
		otel.SetTracerProvider(tp)

		return tp, nil
	}

	exporter, err := newExporter(context.Background(), cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
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

func newExporter(ctx context.Context, cfg config.OTELConfig, logger *slog.Logger) (sdktrace.SpanExporter, error) {
	if cfg.ConsoleExporter {
		logger.InfoContext(ctx, "otel is using stdout trace exporter")

		exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("create stdout trace exporter: %w", err)
		}

		return exporter, nil
	}

	logger.InfoContext(ctx, "otel is using otlp trace exporter", slog.String("endpoint", cfg.Endpoint))

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}

	return exporter, nil
}

func OpenTelemetryModule() fx.Option {
	return fx.Options(
		fx.Provide(NewOpenTelemetryTraceProvider),
	)
}
