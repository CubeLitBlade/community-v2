// Package setup provides the fx dependency injection module for the authz microservice.
package setup

import (
	"context"
	"log/slog"

	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq"
	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/authz"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/authz/transport"
)

// Module returns the fx providers for the authz microservice.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(authz.NewWriter),
		fx.Provide(fx.Annotate(authz.NewChecker, fx.As(new(transport.Checker)))),
		fx.Provide(authz.NewHealthChecker),
		fx.Provide(provideSyncer),
		fx.Provide(
			fx.Annotate(transport.NewHandler, fx.As(new(platform.HTTPMounter)), fx.ResultTags(`group:"mounter"`)),
		),
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
