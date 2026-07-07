package bootstrap

import (
	"context"
	"log/slog"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"go.uber.org/fx"

	v1 "github.com/cubelitblade/community-v2/backend/pkg/contracts/v1"
	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
	"github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq"
	"github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/shared"
)

func provideRabbitMQEnvironment(cfg *shared.RabbitMQConfig, lc fx.Lifecycle) *rmq.Environment {
	env := rmq.NewEnvironment(cfg.BrokerURL, nil)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return env.CloseConnections(ctx)
		},
	})

	return env
}

func providePublisher(env *rmq.Environment, logger *slog.Logger, lc fx.Lifecycle) (*rabbitmq.Publisher, error) {
	publisher := rabbitmq.NewPublisher(env, v1.SystemExchange, logger)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return publisher.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return publisher.Close(ctx)
		},
	})

	return publisher, nil
}

// RabbitMQModule provides the RabbitMQ connection and publisher fx module.
func RabbitMQModule() fx.Option {
	return fx.Options(
		fx.Provide(provideRabbitMQEnvironment),
		fx.Provide(
			fx.Annotate(providePublisher, fx.As(new(outbox.Publisher))),
		),
	)
}
