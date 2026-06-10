package bootstrap

import (
	"context"
	"log/slog"

	"github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"go.uber.org/fx"

	event "github.com/cubelitblade/community-v2/backend/pkg/contracts/v1"
	"github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq"
	account "github.com/cubelitblade/community-v2/backend/services/account/api/events/v1"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/shared"
)

func provideRMQEnv(cfg *shared.RabbitMQConfig, lc fx.Lifecycle) *rabbitmqamqp.Environment {
	env := rabbitmqamqp.NewEnvironment(cfg.BrokerURL, nil)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return env.CloseConnections(ctx)
		},
	})

	return env
}

func provideSubscriber(env *rabbitmqamqp.Environment, logger *slog.Logger, lc fx.Lifecycle) *rabbitmq.QuorumSubscriber {
	subscriber := rabbitmq.NewQuorumSubscriber(env, event.SystemExchange, event.QueueNameAuthz, logger,
		rabbitmq.WithQuorumSubscriberKeys([]string{
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

// RabbitMQModule provides the RabbitMQ connection and subscriber fx module.
func RabbitMQModule() fx.Option {
	return fx.Options(
		fx.Provide(provideRMQEnv),
		fx.Provide(provideSubscriber),
	)
}
