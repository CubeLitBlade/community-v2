package bootstrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const slowThreshold = 200 * time.Millisecond

// DatabaseConfig holds the database connection configuration.
type DatabaseConfig struct {
	DSN string `env:"DATABASE_URL" required:"true"`
}

func gormConfig(appRT AppRuntime) *gorm.Config {
	logLevel := logger.Warn

	if appRT.Environment() == AppEnvProduction {
		logLevel = logger.Warn
	}

	return &gorm.Config{
		Logger: logger.New(log.New(os.Stderr, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             slowThreshold,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  true,
		}),
		TranslateError: true,
	}
}

func provideDatabase(cfg DatabaseConfig, lifecycle fx.Lifecycle, appRT AppRuntime) (*gorm.DB, error) {
	db, sqlDB, err := platform.Open(cfg.DSN, gormConfig(appRT))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return sqlDB.Close()
		},
	})

	return db, nil
}

// DatabaseModule provides the database connection fx module.
func DatabaseModule() fx.Option {
	return fx.Options(
		fx.Provide(provideDatabase),
	)
}
