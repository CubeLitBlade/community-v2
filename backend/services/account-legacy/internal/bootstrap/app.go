package bootstrap

import (
	"net/http"

	"go.uber.org/fx"

	account "github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/account/setup"
	auth "github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/auth/setup"
	health "github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/health/setup"
)

// NewApp creates the fx application with all providers and lifecycle hooks.
func NewApp() *fx.App {
	return fx.New(
		fx.Provide(LoadConfig),

		ConfigModule(),
		CookieModule(),
		JWTModule(),
		SlogModule(),
		GormModule(),
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
