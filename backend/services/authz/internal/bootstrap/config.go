package bootstrap

import (
	"fmt"

	"github.com/caarlos0/env"
	"go.uber.org/fx"
)

// Config is the top-level configuration struct that aggregates all sub-configs.
type Config struct {
	Runtime RuntimeConfig
	Rabbit  RabbitMQConfig
	OpenFGA OpenFGAConfig
}

// LoadConfig loads all configuration from environment variables.
func LoadConfig() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg.Runtime); err != nil {
		return nil, fmt.Errorf("load runtime config: %w", err)
	}

	if err := env.Parse(&cfg.Rabbit); err != nil {
		return nil, fmt.Errorf("load rabbit config: %w", err)
	}

	if err := env.Parse(&cfg.OpenFGA); err != nil {
		return nil, fmt.Errorf("load openfga config: %w", err)
	}

	return &cfg, nil
}

// ConfigModule provides the config loading fx module.
func ConfigModule() fx.Option {
	return fx.Options(
		fx.Provide(LoadConfig),

		fx.Provide(func(cfg *Config) RuntimeConfig { return cfg.Runtime }),
		fx.Provide(func(cfg *Config) RabbitMQConfig { return cfg.Rabbit }),
		fx.Provide(func(cfg *Config) OpenFGAConfig { return cfg.OpenFGA }),

		fx.Provide(
			fx.Annotate(func(cfg *Config) RuntimeConfig { return cfg.Runtime },
				fx.As(new(AppRuntime)),
			),
		),
	)
}
