// Package setup wires the auth module dependencies and registers its HTTP routes.
package setup

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/CubeLitBlade/community-v2/backend/internal/auth"
	"github.com/CubeLitBlade/community-v2/backend/internal/auth/transport"
	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
	"github.com/CubeLitBlade/community-v2/backend/internal/jwt"
)

// Module is the auth domain module, holding its services and HTTP handler.
type Module struct {
	Issuer *jwt.JWT
	Login  *auth.Login

	handler *transport.Handler
}

// ModuleDeps holds the external dependencies needed to initialize the auth module.
type ModuleDeps struct {
	IDs         idgen.Generator
	JWTConfig   *jwt.Config
	AccountAuth auth.AccountAuthenticator
	AuthnMW     func(c *gin.Context)
	Logger      *slog.Logger
}

// NewModule wires the auth domain services and returns a Module.
func NewModule(deps ModuleDeps) *Module {
	issuer := jwt.New(deps.JWTConfig, deps.IDs)
	login := auth.NewLogin(deps.AccountAuth, issuer)

	handler := transport.NewHandler(transport.Deps{
		Login:   login,
		Logger:  deps.Logger,
		AuthnMW: deps.AuthnMW,
	})

	return &Module{
		Issuer:  issuer,
		Login:   login,
		handler: handler,
	}
}

// Mount registers the auth module's HTTP routes on the given router.
func (m *Module) Mount(router gin.IRouter) {
	m.handler.RegisterRoutes(router)
}
