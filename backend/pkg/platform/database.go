// Package platform provides shared infrastructure components for all services:
// database connection management, Gin HTTP engine setup, and request-level middleware.
package platform

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const defaultPingTimeout = 5 * time.Second

// Open connects to a PostgreSQL database, pings to verify connectivity,
// and returns both the GORM instance and the underlying *sql.DB.
func Open(dsn string, cfg *gorm.Config) (gormDB *gorm.DB, sqlDB *sql.DB, err error) {
	gormDB, err = gorm.Open(postgres.Open(dsn), cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err = gormDB.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get sql db: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		pingErr := fmt.Errorf("ping database: %w", err)
		if closeErr := sqlDB.Close(); closeErr != nil {
			return nil, nil, errors.Join(pingErr, fmt.Errorf("close database: %w", closeErr))
		}

		return nil, nil, pingErr
	}

	return gormDB, sqlDB, nil
}
