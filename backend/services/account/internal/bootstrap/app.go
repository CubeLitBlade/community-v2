package bootstrap

import (
	"go.uber.org/fx"
)

func NewApp() *fx.App {
	return fx.New(
		fx.Provide(LoadConfig),

		SlogModule(),
		OpenTelemetryModule(),
		DatabaseModule(),
		SnowflakeModule(),
		GRPCModule(),
		HealthModule(),
		ApplicationModule(),
		RabbitMQModule(),
	)
}
