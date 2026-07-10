package bootstrap

import (
	"context"
	"log/slog"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"go.uber.org/fx"

	v1 "github.com/cubelitblade/community-v2/backend/pkg/contracts/v1"
	"github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq"
	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/config"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence/ent"
)

func provideRabbitMQEnvironment(cfg config.RabbitMQConfig, lc fx.Lifecycle) *rmq.Environment {
	env := rmq.NewEnvironment(cfg.BrokerURL, nil)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return env.CloseConnections(context.WithoutCancel(ctx))
		},
	})

	return env
}

func providePublisher(
	env *rmq.Environment, logger *slog.Logger, lc fx.Lifecycle,
) (persistence.Publisher, error) {
	pub := rabbitmq.NewPublisher(env, v1.SystemExchange, logger)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return pub.Start(context.WithoutCancel(ctx))
		},
		OnStop: func(ctx context.Context) error {
			return pub.Close(context.WithoutCancel(ctx))
		},
	})

	return pub, nil
}

func provideRelay(
	ids idgen.Generator,
	client *ent.Client,
	publisher persistence.Publisher,
	logger *slog.Logger,
	lc fx.Lifecycle,
) *persistence.OutboxRelay {
	relay := persistence.NewOutboxRelay(ids, client, publisher, logger)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			relay.Start(ctx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			relay.Stop()
			return nil
		},
	})

	return relay
}

func RabbitMQModule() fx.Option {
	return fx.Options(
		fx.Provide(provideRabbitMQEnvironment),
		fx.Provide(providePublisher),
		fx.Provide(provideRelay),

		fx.Invoke(func(*persistence.OutboxRelay) {}),
	)
}
