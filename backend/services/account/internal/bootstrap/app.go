package bootstrap

import (
	"net/http"

	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	accountSetup "github.com/cubelitblade/community-v2/backend/services/account/internal/account/setup"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/auth"
	authSetup "github.com/cubelitblade/community-v2/backend/services/account/internal/auth/setup"
	healthSetup "github.com/cubelitblade/community-v2/backend/services/account/internal/health/setup"
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

		fx.Provide(
			fx.Annotate(accountSetup.NewAuthenticatorAdapter, fx.As(new(auth.AccountAuthenticator))),
		),
		fx.Provide(
			fx.Annotate(accountSetup.NewLoginRecorder, fx.As(new(auth.LoginRecorder))),
		),
		fx.Invoke(registerAPIRoutes),
		fx.Invoke(registerHealthRoutes),
		fx.Invoke(func(*http.Server) {}),
	)
}
