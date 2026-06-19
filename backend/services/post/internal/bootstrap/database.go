package bootstrap

import (
	"context"
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/XSAM/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence/ent"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/config"
)

func NewDatabase(lc fx.Lifecycle, cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := otelsql.Open("pgx", cfg.URI, otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return db.PingContext(ctx)
		},
		OnStop: func(context.Context) error {
			return db.Close()
		},
	})

	return db, nil
}

func NewEntClient(db *sql.DB) *ent.Client {
	return ent.NewClient(
		ent.Driver(entsql.OpenDB(dialect.Postgres, db)),
	)
}

func DBModule() fx.Option {
	return fx.Options(
		fx.Provide(NewDatabase),
		fx.Provide(NewEntClient),

		fx.Provide(persistence.NewEntTransactor),
		fx.Provide(persistence.NewPostSaver),
	)
}
