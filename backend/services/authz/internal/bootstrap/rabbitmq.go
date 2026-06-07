package bootstrap

import (
	"context"
	"log/slog"

	"github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq"
	v1 "github.com/cubelitblade/community-v2/backend/services/account/api/events/v1"
	"github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"go.uber.org/fx"
)

// RabbitMQConfig holds the RabbitMQ connection and topology configuration.
type RabbitMQConfig struct {
	BrokerURL    string `env:"RMQ_URL" required:"true"`
	ExchangeName string `env:"RMQ_SYSTEM_EXCHANGE_NAME" required:"true"`
	QueueName    string `env:"RMQ_QUEUE_NAME_AUTHZ" required:"true"`
}

func provideRMQEnv(cfg RabbitMQConfig, lc fx.Lifecycle) *rabbitmqamqp.Environment {
	env := rabbitmqamqp.NewEnvironment(cfg.BrokerURL, nil)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return env.CloseConnections(ctx)
		},
	})

	return env
}

func provideSubscriber(
	cfg RabbitMQConfig, env *rabbitmqamqp.Environment, logger *slog.Logger, lc fx.Lifecycle,
) *rabbitmq.QuorumSubscriber {
	subscriber := rabbitmq.NewQuorumSubscriber(env, cfg.ExchangeName, cfg.QueueName, logger,
		rabbitmq.WithQuorumSubscriberKeys([]string{
			v1.TopicAccountCreated,
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
