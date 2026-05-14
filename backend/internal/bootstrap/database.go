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

func OpenDatabase(dsn string) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: NewGormLogger(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get database handle: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), databasePingTimeout)
	defer cancel()

	err = sqlDB.PingContext(ctx)
	if err != nil {
		pingErr := fmt.Errorf("ping database: %w", err)

		closeErr := sqlDB.Close()
		if closeErr != nil {
			return nil, nil, errors.Join(
				pingErr,
				fmt.Errorf("close database: %w", closeErr),
			)
		}

		return nil, nil, pingErr
	}

	return db, sqlDB, nil
}
