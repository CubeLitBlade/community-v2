package bootstrap

import (
	"go.uber.org/fx"
)

// NewApp creates the fx application with all providers and lifecycle hooks.
func NewApp() *fx.App {
	return fx.New(
		fx.Provide(LoadConfig),

		ApplicationModule(),
		GRPCModule(),
		HealthModule(),
		OpenFGAModule(),
		RabbitMQModule(),
		SlogModule(),
		OpenTelemetryModule(),
		RedisModule(),
	)
}
