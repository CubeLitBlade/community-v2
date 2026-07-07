package bootstrap

import (
	"context"
	"log/slog"

	"github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"go.uber.org/fx"

	event "github.com/cubelitblade/community-v2/backend/pkg/contracts/v1"
	rmq "github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq"
	account "github.com/cubelitblade/community-v2/backend/services/account-legacy/api/events/v1"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/adapter/driven/rabbitmq"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/application"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/config"
)

func NewRabbitMQEnv(cfg config.RabbitMQConfig, lc fx.Lifecycle) *rabbitmqamqp.Environment {
	env := rabbitmqamqp.NewEnvironment(cfg.BrokerURL, nil)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return env.CloseConnections(ctx)
		},
	})

	return env
}

func NewRabbitMQSubscriber(
	env *rabbitmqamqp.Environment, logger *slog.Logger, lc fx.Lifecycle,
) *rmq.QuorumSubscriber {
	subscriber := rmq.NewQuorumSubscriber(env, event.SystemExchange, event.QueueNameAuthz, logger,
		rmq.WithQuorumSubscriberKeys([]string{
			account.TopicAccountCreated,
		}),
	)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return subscriber.Close(ctx)
		},
	})

	return subscriber
}

func NewRabbitMQConsumer(
	lc fx.Lifecycle, subscriber *rmq.QuorumSubscriber, syncer *application.EventSyncer, logger *slog.Logger,
) *rabbitmq.Consumer {
	consumer := rabbitmq.NewConsumer(subscriber, syncer, logger)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return consumer.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			consumer.Stop(ctx)
			return nil
		},
	})

	return consumer
}

// RabbitMQModule provides the RabbitMQ connection and subscriber fx module.
func RabbitMQModule() fx.Option {
	return fx.Options(
		fx.Provide(NewRabbitMQEnv),
		fx.Provide(NewRabbitMQSubscriber),
		fx.Provide(NewRabbitMQConsumer),

		fx.Invoke(func(*rabbitmq.Consumer) {}),
	)
}
