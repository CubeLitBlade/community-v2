// Package setup provides constructors that wire together account domain services, persistence, and HTTP transport.
package setup

import (
	"context"
	"log/slog"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/account"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/account/storage"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/account/transport"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/auth"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

// Module returns the fx providers for the account module.
// Cross-module interface bindings (AccountAuthenticator, LoginRecorder)
// are handled by the composition root in bootstrap.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				storage.NewAccountWriter,
				fx.As(new(account.Creator)),
				fx.As(new(account.LastLoginUpdater)),
			),
		),
		fx.Provide(
			fx.Annotate(
				storage.NewAccountReader,
				fx.As(new(account.ByIDFinder)),
				fx.As(new(account.ByUsernameFinder)),
			),
		),
		fx.Provide(
			fx.Annotate(
				storage.NewOutboxRepository,
				fx.As(new(account.OutboxSaver)),
				fx.As(new(outbox.Scanner)),
				fx.As(new(outbox.Recorder)),
			),
		),
		fx.Provide(provideOutboxRelay),
		fx.Provide(
			fx.Annotate(account.NewRegistrar, fx.As(new(transport.Registrar))),
		),
		fx.Provide(
			fx.Annotate(account.NewAuthenticator, fx.As(new(auth.Authenticator))),
		),
		fx.Provide(
			fx.Annotate(account.NewProfileFinder, fx.As(new(transport.ProfileFinder))),
		),
		fx.Provide(
			fx.Annotate(account.NewLoginRecorder, fx.As(new(auth.LoginRecorder))),
		),
		fx.Provide(
			fx.Annotate(transport.NewHandler, fx.As(new(platform.HTTPMounter)), fx.ResultTags(`group:"mounter"`)),
		),
		fx.Invoke(func(*outbox.Relay) {}),
	)
}

func provideOutboxRelay(
	ids idgen.Generator, scanner outbox.Scanner, recorder outbox.Recorder,
	publisher outbox.Publisher, db *gorm.DB, logger *slog.Logger, lc fx.Lifecycle,
) *outbox.Relay {
	relay := outbox.NewRelay(
		ids, scanner, recorder, publisher, db, logger,
	)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				relay.Start(context.WithoutCancel(ctx))
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return relay.Close()
		},
	})

	return relay
}
