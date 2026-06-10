package bootstrap

import (
	"fmt"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/shared"
	"github.com/go-playground/validator/v10"
	toml "github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"
)

// LoadConfig reads and validates the account service configuration from the
// TOML file located at services/account/config.toml.
func LoadConfig() (*shared.Config, error) {
	loader := koanf.New(".")

	if err := loader.Load(file.Provider("services/account/config.toml"), toml.Parser()); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	var cfg shared.Config
	if err := loader.Unmarshal("account", &cfg); err != nil {
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
		fx.Provide(func(cfg *shared.Config) *shared.DatabaseConfig { return cfg.Database }),
		fx.Provide(func(cfg *shared.Config) *shared.GormConfig { return cfg.Gorm }),
		fx.Provide(func(cfg *shared.Config) *shared.GinConfig { return cfg.Gin }),
		fx.Provide(func(cfg *shared.Config) *shared.SlogConfig { return cfg.Slog }),
		fx.Provide(func(cfg *shared.Config) *shared.SnowflakeConfig { return cfg.Snowflake }),
		fx.Provide(func(cfg *shared.Config) *shared.RabbitMQConfig { return cfg.RabbitMQ }),
		fx.Provide(func(cfg *shared.Config) *shared.JWTConfig { return cfg.JWT }),
		fx.Provide(func(cfg *shared.Config) *shared.CookieConfig { return cfg.Cookie }),
		fx.Provide(func(cfg *shared.Config) *shared.TokenConfig { return cfg.Token }),
	)
}
