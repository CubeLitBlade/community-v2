package bootstrap

import (
	"go.uber.org/fx"
)

func NewApp() *fx.App {
	return fx.New(
		fx.Provide(LoadConfig),

		ApplicationModule(),
		AuthzModule(),
		DBModule(),
		GRPCModule(),
		SlogModule(),
		SnowflakeModule(),
		OpenTelemetryModule(),
		RabbitMQModule(),
	)
}
