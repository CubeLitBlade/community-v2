// Package setup provides constructors that wire together auth domain services,
// JWT issuance, and HTTP transport.
package setup

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/CubeLitBlade/community-v2/backend/internal/auth"
	"github.com/CubeLitBlade/community-v2/backend/internal/auth/transport"
	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
	"github.com/CubeLitBlade/community-v2/backend/internal/jwt"
)

// NewHandler creates the auth HTTP handler with all dependencies wired.
func NewHandler(
	ids idgen.Generator,
	jwtCfg *jwt.Config,
	accAuth auth.AccountAuthenticator,
	recorder auth.LoginRecorder,
	authnMW func(*gin.Context),
	logger *slog.Logger,
) *transport.Handler {
	issuer := jwt.New(jwtCfg, ids)
	login := auth.NewLogin(accAuth, issuer, recorder)

	return transport.NewHandler(transport.Deps{
		Login:   login,
		Logger:  logger,
		AuthnMW: authnMW,
	})
}
