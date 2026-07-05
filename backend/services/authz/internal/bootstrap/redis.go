package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	redismetrics "github.com/redis/go-redis/extra/redisotel-native/v9"
	"github.com/redis/go-redis/extra/redisotel/v9"
	redisdk "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/adapter/driven/redis"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/config"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/telemetry"
)

func initRedisOTEL() error {
	cfg := redismetrics.NewConfig().
		WithEnabled(true).
		WithMetricGroups(redismetrics.MetricGroupFlagCommand |
			redismetrics.MetricGroupFlagConnectionBasic |
			redismetrics.MetricGroupFlagResiliency |
			redismetrics.MetricGroupFlagConnectionAdvanced).
		WithIncludeCommands([]string{"GET", "SET"}).
		WithExcludeCommands([]string{"DEBUG", "SLOWLOG"}).
		WithHidePubSubChannelNames(true).
		WithHideStreamNames(true).
		WithHistogramBuckets([]float64{
			0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0, 10.0,
		})

	if err := redismetrics.GetObservabilityInstance().Init(cfg); err != nil {
		return fmt.Errorf("init redis otel: %w", err)
	}

	return nil
}

func newRedisHooks(
	rdb *redisdk.Client, logger *slog.Logger,
) fx.Hook {
	return fx.Hook{
		OnStart: func(ctx context.Context) error {
			_, err := rdb.Ping(ctx).Result()
			if err != nil {
				return fmt.Errorf("failed to ping redis: %w", err)
			}

			return nil
		},
		OnStop: func(context.Context) error {
			if err := rdb.Close(); err != nil {
				logger.Error("failed to close redis client", slog.Any("error", err))
			}

			if err := redismetrics.GetObservabilityInstance().Shutdown(); err != nil {
				logger.Error("failed to shut down redis otel instance", slog.Any("error", err))
			}

			logger.Info("redis client stopped gracefully")

			return nil
		},
	}
}

func NewRedisClient(
	lc fx.Lifecycle, cfg config.RedisConfig, logger *slog.Logger, _ metric.MeterProvider,
) (*redisdk.Client, error) {
	if err := initRedisOTEL(); err != nil {
		return nil, fmt.Errorf("init redis client: %w", err)
	}

	rdb := redisdk.NewClient(&redisdk.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := redisotel.InstrumentTracing(rdb); err != nil {
		return nil, fmt.Errorf("failed to instrument tracing: %w", err)
	}

	lc.Append(newRedisHooks(rdb, logger))

	logger.Info("redis, set, go!")

	return rdb, nil
}

func RedisModule() fx.Option {
	return fx.Options(
		fx.Provide(NewRedisClient),
		fx.Provide(redis.NewChecker),
		fx.Provide(redis.NewWriter),
		fx.Provide(redis.NewInvalidator),

		fx.Decorate(telemetry.NewInstrumentedCacheChecker),
		fx.Decorate(telemetry.NewInstrumentedCacheWriter),
		fx.Decorate(telemetry.NewInstrumentedCacheInvalidator),
	)
}
