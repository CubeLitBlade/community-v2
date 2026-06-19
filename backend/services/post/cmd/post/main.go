// package main starts the community-v2 service
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/bootstrap"
	"github.com/cubelitblade/community-v2/backend/services/post/migrations"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "Run database migrations.")

	flag.Parse()

	if *migrateOnly {
		if err := runMigrations(); err != nil {
			//nolint:sloglint // CLI mode: custom logger is not bootstrapped, using default is intended.
			slog.Error("migration failed", slog.Any("error", err))
			os.Exit(1)
		}

		return
	}

	bootstrap.NewApp().Run()
}

func runMigrations() error {
	uri := os.Getenv("POST_MIGRATE_URI")

	db, err := sql.Open("pgx", uri)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	//nolint:sloglint // CLI mode: custom logger is not bootstrapped, using default is intended.
	slog.Info("running database migrations...")

	if err := migrations.Up(db); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	//nolint:sloglint // CLI mode: custom logger is not bootstrapped, using default is intended.
	slog.Info("migrations completed successfully")

	return nil
}
