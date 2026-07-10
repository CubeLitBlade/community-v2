package bootstrap

import (
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/application"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/port"
)

func ApplicationModule() fx.Option {
	return fx.Options(
		fx.Provide(persistence.NewOutboxRecorder),
		fx.Provide(func(recorder *persistence.OutboxRecorder) port.EventRecorder {
			return recorder
		}),
		fx.Provide(application.NewPublisher),
	)
}
