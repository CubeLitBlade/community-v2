// Package platform provides shared infrastructure (DB, HTTP, logging, Gin,
// Snowflake, JWT) wired as fx constructors for use by service modules.
package platform

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const databasePingTimeout = 5 * time.Second

// DBConfig provides the Data Source Name for the database connection.
type DBConfig struct {
	DSN string
}

// OpenDatabase connects to a PostgreSQL database, pings to verify connectivity,
// and returns both the GORM instance and the underlying *sql.DB.
func OpenDatabase(dsn string, gormLogger logger.Interface) (gormDB *gorm.DB, sqlDB *sql.DB, err error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         gormLogger,
		TranslateError: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	handle, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get database handle: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), databasePingTimeout)
	defer cancel()

	if err := handle.PingContext(ctx); err != nil {
		pingErr := fmt.Errorf("ping database: %w", err)
		if closeErr := handle.Close(); closeErr != nil {
			return nil, nil, errors.Join(pingErr, fmt.Errorf("close database: %w", closeErr))
		}
		return nil, nil, pingErr
	}

	return db, handle, nil
}

// ProvideDatabase creates a *gorm.DB, registers a close hook for the
// connection pool, and returns only the *gorm.DB.
func ProvideDatabase(lifecycle fx.Lifecycle, cfg *DBConfig, gormLogger logger.Interface) (*gorm.DB, error) {
	db, sqlDB, err := OpenDatabase(cfg.DSN, gormLogger)
	if err != nil {
		return nil, err
	}

	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return sqlDB.Close()
		},
	})

	return db, nil
}
