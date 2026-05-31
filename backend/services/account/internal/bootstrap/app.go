package bootstrap

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
	"github.com/cubelitblade/community-v2/backend/pkg/events/rabbitmq"
	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/account"
	accountSetup "github.com/cubelitblade/community-v2/backend/services/account/internal/account/setup"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/auth"
	authSetup "github.com/cubelitblade/community-v2/backend/services/account/internal/auth/setup"
	healthSetup "github.com/cubelitblade/community-v2/backend/services/account/internal/health/setup"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"go.uber.org/fx"
)

// NewApp creates the fx application with all providers and lifecycle hooks.
func NewApp() *fx.App {
	return fx.New(
		fx.Provide(ProvideConfig),
		platform.StandardModules(),

		accountSetup.Module(),
		authSetup.Module(),
		healthSetup.Module(),

		fx.Provide(provideRabbitMQEnv),
		fx.Provide(
			fx.Annotate(provideEventsPublisher, fx.As(new(outbox.Publisher))),
		),

		fx.Provide(
			fx.Annotate(accountSetup.NewAuthenticatorAdapter, fx.As(new(auth.AccountAuthenticator))),
		),
		fx.Provide(
			fx.Annotate(account.NewLoginRecorder, fx.As(new(auth.LoginRecorder))),
		),
		fx.Invoke(registerAPIRoutes),
		fx.Invoke(registerHealthRoutes),
		fx.Invoke(func(*http.Server) {}),
	)
}

func provideRabbitMQEnv(cfg *RabbitMQConfig, lc fx.Lifecycle) *rmq.Environment {
	env := rmq.NewEnvironment(cfg.BrokerURL, nil)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error { return env.CloseConnections(ctx) },
	})

	return env
}

func provideEventsPublisher(
	env *rmq.Environment, cfg *RabbitMQConfig, logger *slog.Logger, lc fx.Lifecycle,
) *rabbitmq.Publisher {
	opts := []rabbitmq.PublisherOption{
		rabbitmq.WithPublisherLogger(logger),
	}

	pub := rabbitmq.NewPublisher(env, cfg.ExchangeName, opts...)

	lc.Append(fx.Hook{
		OnStart: pub.Start,
		OnStop:  func(ctx context.Context) error { return pub.Close(ctx) },
	})

	return pub
}
