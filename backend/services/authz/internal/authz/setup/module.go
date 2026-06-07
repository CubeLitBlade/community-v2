// Package setup provides the fx dependency injection module for the authz microservice.
package setup

import (
	"context"
	"log/slog"

	"github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/authz"
	"go.uber.org/fx"
)

// Module returns the fx providers for the authz microservice.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(authz.NewWriter),
		fx.Provide(authz.NewChecker),
		fx.Provide(authz.NewHealthChecker),
		fx.Provide(provideSyncer),
		fx.Invoke(func(*authz.Syncer) {}),
	)
}

func provideSyncer(
	subscriber *rabbitmq.QuorumSubscriber, writer *authz.Writer, logger *slog.Logger, lc fx.Lifecycle,
) *authz.Syncer {
	syncer := authz.NewSyncer(subscriber, writer, logger)

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				return syncer.Start(context.WithoutCancel(ctx))
			},
			OnStop: func(ctx context.Context) error {
				syncer.Stop(ctx)
				return nil
			},
		},
	)

	return syncer
}
