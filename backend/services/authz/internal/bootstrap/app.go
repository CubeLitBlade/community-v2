package bootstrap

import (
	"net/http"

	authz "github.com/cubelitblade/community-v2/backend/services/authz/internal/authz/setup"
	health "github.com/cubelitblade/community-v2/backend/services/authz/internal/health/setup"
	"go.uber.org/fx"
)

// NewApp creates the fx application with all providers and lifecycle hooks.
func NewApp() *fx.App {
	return fx.New(
		ConfigModule(),
		RabbitMQModule(),
		LoggerModule(),
		FGAModule(),
		GinModule(),
		HTTPServer(),

		authz.Module(),
		health.Module(),

		fx.Invoke(registerAPIRoutes),
		fx.Invoke(registerHealthRoutes),
		fx.Invoke(func(*http.Server) {}),
	)
}
