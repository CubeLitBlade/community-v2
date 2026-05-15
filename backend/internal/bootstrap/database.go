package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const databasePingTimeout = 5 * time.Second

// OpenDatabase connects to the database via DSN, pings to verify connectivity,
// and returns both the GORM and standard database/sql DB instances.
func OpenDatabase(dsn string) (gormDB *gorm.DB, sqlDB *sql.DB, err error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: NewGormLogger(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	handle, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get database handle: %w", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		databasePingTimeout,
	)
	defer cancel()

	err = handle.PingContext(ctx)
	if err != nil {
		pingErr := fmt.Errorf("ping database: %w", err)

		closeErr := handle.Close()
		if closeErr != nil {
			return nil, nil, errors.Join(
				pingErr,
				fmt.Errorf("close database: %w", closeErr),
			)
		}

		return nil, nil, pingErr
	}

	return db, handle, nil
}
