package bootstrap

import (
	"net/http"

	account "github.com/cubelitblade/community-v2/backend/services/account/internal/account/setup"
	auth "github.com/cubelitblade/community-v2/backend/services/account/internal/auth/setup"
	health "github.com/cubelitblade/community-v2/backend/services/account/internal/health/setup"
	"go.uber.org/fx"
)

// NewApp creates the fx application with all providers and lifecycle hooks.
func NewApp() *fx.App {
	return fx.New(
		ConfigModule(),
		CookieModule(),
		JWTModule(),
		LoggerModule(),
		DatabaseModule(),
		GinModule(),
		SnowflakeModule(),
		RabbitMQModule(),
		HTTPServer(),

		account.Module(),
		auth.Module(),
		health.Module(),

		fx.Invoke(registerAPIRoutes),
		fx.Invoke(registerHealthRoutes),
		fx.Invoke(func(*http.Server) {}),
	)
}
