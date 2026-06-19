package persistence_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence/ent"
	"github.com/cubelitblade/community-v2/backend/services/post/migrations"
)

//nolint:unused // Actually used.
var (
	authorID    = int64(1)
	username    = "alice"
	displayName = "Alice"

	postID  = int64(2)
	title   = "Lorem Ipsum"
	content = "Lorem ipsum dolor sit amet, consectetur adipiscing elit."

	fixedTime = time.Date(2026, 6, 13, 0, 0, 0, 0, time.UTC)
)

//nolint:unused // Actually used.
func newClient(ctx context.Context, t *testing.T) *ent.Client {
	t.Helper()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_password"),
		postgres.BasicWaitStrategies(),
	)

	t.Cleanup(func() {
		testcontainers.CleanupContainer(t, pgContainer)
	})

	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable&timezone=UTC")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	t.Cleanup(func() {
		client.Close()
	})

	if err := migrations.Up(drv.DB()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return client
}
