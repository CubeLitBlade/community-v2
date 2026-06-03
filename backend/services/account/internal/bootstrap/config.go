// Package bootstrap is the composition root for the account service. It owns
// config loading, fx app construction, and module wiring.
package bootstrap

import (
	"fmt"

	"github.com/caarlos0/env"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/contracts"
	"go.uber.org/fx"
)

// Config is the top-level configuration struct that aggregates all sub-configs.
type Config struct {
	Runtime   RuntimeConfig
	Database  DatabaseConfig
	RabbitMQ  RabbitMQConfig
	Snowflake SnowflakeConfig
	JWT       JWTConfig
	Token     TokenConfig
}

// LoadConfig loads all configuration from environment variables.
func LoadConfig() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg.Runtime); err != nil {
		return nil, fmt.Errorf("load runtime config: %w", err)
	}

	if err := env.Parse(&cfg.Database); err != nil {
		return nil, fmt.Errorf("load database config: %w", err)
	}

	if err := env.Parse(&cfg.RabbitMQ); err != nil {
		return nil, fmt.Errorf("load rabbitmq config: %w", err)
	}

	if err := env.Parse(&cfg.Snowflake); err != nil {
		return nil, fmt.Errorf("load snowflake config: %w", err)
	}

	if err := env.Parse(&cfg.JWT); err != nil {
		return nil, fmt.Errorf("load jwt config: %w", err)
	}

	if err := env.Parse(&cfg.Token); err != nil {
		return nil, fmt.Errorf("load token config: %w", err)
	}

	return &cfg, nil
}

// ConfigModule provides the config loading fx module.
func ConfigModule() fx.Option {
	return fx.Options(
		fx.Provide(LoadConfig),

		fx.Provide(func(cfg *Config) SnowflakeConfig { return cfg.Snowflake }),
		fx.Provide(func(cfg *Config) RabbitMQConfig { return cfg.RabbitMQ }),
		fx.Provide(func(cfg *Config) DatabaseConfig { return cfg.Database }),
		fx.Provide(func(cfg *Config) JWTConfig { return cfg.JWT }),

		fx.Provide(
			fx.Annotate(func(cfg *Config) RuntimeConfig { return cfg.Runtime },
				fx.As(new(AppRuntime)),
			),
		),
		fx.Provide(
			fx.Annotate(func(cfg *Config) TokenConfig { return cfg.Token },
				fx.As(new(contracts.TTLProvider)),
				fx.As(new(IssuerProvider)),
			),
		),
	)
}
