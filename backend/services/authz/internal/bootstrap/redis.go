package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/extra/redisotel-native/v9"
	redisdk "github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/adapter/driven/redis"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/config"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

func NewRedisClient(lc fx.Lifecycle, cfg config.RedisConfig, logger *slog.Logger) (*redisdk.Client, error) {
	otelInstance := redisotel.GetObservabilityInstance()

	otelcfg := redisotel.NewConfig().WithEnabled(true)
	if err := otelInstance.Init(otelcfg); err != nil {
		return nil, fmt.Errorf("init redis client: %w", err)
	}

	rdb := redisdk.NewClient(&redisdk.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	lc.Append(fx.Hook{
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

			logger.Info("redis client stopped gracefully")

			return nil
		},
	})

	logger.Info("redis, set, go!")

	return rdb, nil
}

func RedisModule() fx.Option {
	return fx.Options(
		fx.Provide(NewRedisClient),
		fx.Provide(redis.NewChecker),
		fx.Provide(redis.NewWriter),
		fx.Provide(redis.NewInvalidator),

		fx.Invoke(func(port.CacheWriter) {}),
		fx.Invoke(func(port.CacheChecker) {}),
		fx.Invoke(func(port.CacheInvalidator) {}),
	)
}
