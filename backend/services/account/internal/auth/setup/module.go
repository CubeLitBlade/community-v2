// Package setup provides constructors that wire together auth domain services,
// JWT issuance, and HTTP transport.
package setup

import (
	"log/slog"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/pkg/common/jwt"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/auth"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/auth/transport"
	"go.uber.org/fx"
)

// Module returns the fx providers for the auth module.
// Handler annotations (HTTPMounter) are handled by the composition root in bootstrap.
func Module() fx.Option {
	return fx.Options()
}

// NewHandler creates the auth HTTP handler with all dependencies wired.
func NewHandler(
	ids idgen.Generator,
	jwtCfg *jwt.Config,
	cookieCfg transport.CookieConfig,
	accAuth auth.AccountAuthenticator,
	recorder auth.LoginRecorder,
	logger *slog.Logger,
) *transport.Handler {
	issuer := jwt.New(jwtCfg, ids)
	login := auth.NewLogin(accAuth, issuer, recorder)

	return transport.NewHandler(transport.Deps{
		Login:     login,
		CookieCfg: cookieCfg,
		Logger:    logger,
	})
}
