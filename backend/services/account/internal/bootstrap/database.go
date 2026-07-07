package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/config"
)

type Database struct {
	fx.Out

	GormDB *gorm.DB
	SqlDB  *sql.DB
}

func provideDatabase(dbCfg config.DatabaseConfig, logger *slog.Logger, lc fx.Lifecycle) (Database, error) {
	gormDB, sqlDB, err := platform.Open(dbCfg.URL, &gorm.Config{})
	if err != nil {
		return Database{}, fmt.Errorf("open database: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.InfoContext(ctx, "closing database connection")
			return sqlDB.Close()
		},
	})

	logger.InfoContext(context.Background(), "database connection established")

	return Database{GormDB: gormDB, SqlDB: sqlDB}, nil
}

func DatabaseModule() fx.Option {
	return fx.Options(
		fx.Provide(provideDatabase),
	)
}
