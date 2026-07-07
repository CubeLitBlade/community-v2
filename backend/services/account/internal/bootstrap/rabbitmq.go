package bootstrap

import (
	"context"
	"log/slog"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"go.uber.org/fx"

	v1 "github.com/cubelitblade/community-v2/backend/pkg/contracts/v1"
	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
	"github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq"
	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/config"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/adapter/driven/persistence"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/telemetry"
	"gorm.io/gorm"
)

func provideRabbitMQEnvironment(cfg config.RabbitMQConfig, lc fx.Lifecycle) *rmq.Environment {
	env := rmq.NewEnvironment(cfg.BrokerURL, nil)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return env.CloseConnections(ctx)
		},
	})

	return env
}

func providePublisher(
	env *rmq.Environment, logger *slog.Logger, tel *telemetry.Telemetry, lc fx.Lifecycle,
) (outbox.Publisher, error) {
	inner := rabbitmq.NewPublisher(env, v1.SystemExchange, logger)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return inner.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return inner.Close(ctx)
		},
	})

	return telemetry.NewInstrumentedPublisher(inner, tel), nil
}

func provideRelay(
	ids idgen.Generator,
	outboxRepo *persistence.OutboxRepository,
	publisher outbox.Publisher,
	db *gorm.DB,
	logger *slog.Logger,
	lc fx.Lifecycle,
) *outbox.Relay {
	relay := outbox.NewRelay(ids, outboxRepo, outboxRepo, publisher, db, logger)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			relay.Start(ctx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return relay.Close()
		},
	})

	return relay
}

func RabbitMQModule() fx.Option {
	return fx.Options(
		fx.Provide(provideRabbitMQEnvironment),
		fx.Provide(providePublisher),
		fx.Provide(provideRelay),

		fx.Invoke(func(*outbox.Relay) {}),
	)
}
