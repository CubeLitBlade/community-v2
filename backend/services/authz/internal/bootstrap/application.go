package bootstrap

import (
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/application"
)

func ApplicationModule() fx.Option {
	return fx.Options(
		fx.Provide(application.NewAuthorizer),
		fx.Provide(application.NewEventSyncer),
		fx.Provide(application.NewTupleWriter),
		fx.Provide(fx.Annotate(application.NewHealthChecker, fx.ParamTags(`group:"health_checkers"`))),
	)
}
