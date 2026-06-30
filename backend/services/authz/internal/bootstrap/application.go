package bootstrap

import (
	"context"

	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/application"
)

func ApplicationModule() fx.Option {
	return fx.Options(
		fx.Provide(application.NewAuthorizer),
		fx.Provide(application.NewEventSyncer),
		fx.Provide(
			fx.Annotate(application.NewTupleWriter, fx.OnStop(
				func(ctx context.Context, writer *application.TupleWriter) error {
					return writer.Stop(ctx)
				}))),
		fx.Provide(fx.Annotate(application.NewHealthChecker, fx.ParamTags(`group:"health_checkers"`))),
	)
}
