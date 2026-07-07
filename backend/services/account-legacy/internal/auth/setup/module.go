// Package setup provides constructors that wire together auth domain services,
// JWT issuance, and HTTP transport.
package setup

import (
	"log/slog"

	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/pkg/common/jwt"
	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	"github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/auth"
	"github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/auth/transport"
	"github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/shared"
)

func provideLogin(
	authenticator auth.Authenticator, signer *jwt.Signer, recorder auth.LoginRecorder, cfg *shared.TokenConfig,
) *auth.Login {
	return auth.NewLogin(authenticator, signer, recorder, cfg.AccessTokenValidity)
}

func provideHandler(
	login *auth.Login, logger *slog.Logger, cfg *shared.TokenConfig, policy transport.CookiePolicy,
) *transport.Handler {
	return transport.NewHandler(login, logger, cfg.AccessTokenValidity, policy)
}

// Module returns the fx providers for the auth module.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(provideLogin),
		fx.Provide(
			fx.Annotate(provideHandler, fx.As(new(platform.HTTPMounter)), fx.ResultTags(`group:"mounter"`)),
		),
	)
}
