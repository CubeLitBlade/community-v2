package bootstrap

import (
	"context"
	"log/slog"

	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
	"github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"go.uber.org/fx"
)

// RabbitMQConfig holds the RabbitMQ connection and topology configuration.
type RabbitMQConfig struct {
	BrokerURL    string `env:"RMQ_URL" required:"true"`
	ExchangeName string `env:"RMQ_SYSTEM_EXCHANGE_NAME" required:"true"`
}

func provideRabbitMQEnvironment(cfg RabbitMQConfig, lc fx.Lifecycle) *rmq.Environment {
	env := rmq.NewEnvironment(cfg.BrokerURL, nil)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return env.CloseConnections(ctx)
		},
	})

	return env
}

func providePublisher(
	cfg RabbitMQConfig, env *rmq.Environment, logger *slog.Logger, lc fx.Lifecycle,
) (*rabbitmq.Publisher, error) {
	publisher := rabbitmq.NewPublisher(env, cfg.ExchangeName, logger)

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
