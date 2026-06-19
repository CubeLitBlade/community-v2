package bootstrap

import (
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/application"
)

func ApplicationModule() fx.Option {
	return fx.Options(
		fx.Provide(application.NewPublisher),
	)
}
