package bootstrap

import (
	"net/http"

	"go.uber.org/fx"

	authz "github.com/cubelitblade/community-v2/backend/services/authz/internal/authz/setup"
	health "github.com/cubelitblade/community-v2/backend/services/authz/internal/health/setup"
)

// NewApp creates the fx application with all providers and lifecycle hooks.
func NewApp() *fx.App {
	return fx.New(
		ConfigModule(),
		RabbitMQModule(),
		SlogModule(),
		OpenFGAModule(),
		GinModule(),
		HTTPServer(),

		authz.Module(),
		health.Module(),

		fx.Invoke(registerAPIRoutes),
		fx.Invoke(registerHealthRoutes),
		fx.Invoke(func(*http.Server) {}),
	)
}
