package bootstrap

import (
	"context"
	"log/slog"

	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/application"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/telemetry"
)

func ApplicationModule() fx.Option {
	return fx.Options(
		fx.Provide(func(
			authChecker port.AuthorityChecker,
			cacheChecker port.CacheChecker,
			cacheWriter port.CacheWriter,
			logger *slog.Logger,
			tel *telemetry.Telemetry,
		) application.AuthChecker {
			authz := application.NewAuthorizer(authChecker, cacheChecker, cacheWriter, logger)
			return telemetry.NewInstrumentedAuthorizer(authz, tel, logger)
		}),
		fx.Provide(application.NewEventSyncer),
		fx.Provide(
			fx.Annotate(application.NewTupleWriter, fx.OnStop(
				func(ctx context.Context, writer *application.TupleWriter) error {
					return writer.Stop(ctx)
				}))),
		fx.Provide(fx.Annotate(application.NewHealthChecker, fx.ParamTags(`group:"health_checkers"`))),
	)
}
