package bootstrap

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/shared"
)

// LoadConfig reads and validates the authz service configuration from the
// TOML file located at services/authz/config.toml.
func LoadConfig() (*shared.Config, error) {
	loader := koanf.New(".")

	if err := loader.Load(file.Provider("services/authz/config.toml"), toml.Parser()); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	var cfg shared.Config
	if err := loader.Unmarshal("authz", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

// ConfigModule provides an fx.Option that supplies all sub-configurations as individually injectable dependencies.
func ConfigModule() fx.Option {
	return fx.Options(
		fx.Provide(LoadConfig),

		fx.Provide(func(cfg *shared.Config) *shared.OpenFGAConfig { return cfg.OpenFGA }),
		fx.Provide(func(cfg *shared.Config) *shared.RabbitMQConfig { return cfg.RabbitMQ }),
		fx.Provide(func(cfg *shared.Config) *shared.GinConfig { return cfg.Gin }),
		fx.Provide(func(cfg *shared.Config) *shared.SlogConfig { return cfg.Slog }),
	)
}
