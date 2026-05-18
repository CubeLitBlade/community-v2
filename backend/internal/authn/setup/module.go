// Package setup wires the authn module dependencies and registers its HTTP routes.
package setup

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/CubeLitBlade/community-v2/backend/internal/authn"
	"github.com/CubeLitBlade/community-v2/backend/internal/authn/transport"
	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
	"github.com/CubeLitBlade/community-v2/backend/internal/jwt"
)

// Module is the authn domain module, holding its services and HTTP handler.
type Module struct {
	Issuer *jwt.Issuer
	Login  *authn.Login

	handler *transport.Handler
}

// ModuleDeps holds the external dependencies needed to initialize the authn module.
type ModuleDeps struct {
	IDs         idgen.Generator
	JWTConfig   *jwt.Config
	AccountAuth authn.AccountAuthenticator
	Logger      *slog.Logger
}

// NewModule wires the authn domain services and returns a Module.
func NewModule(deps ModuleDeps) *Module {
	issuer := jwt.NewIssuer(deps.JWTConfig, deps.IDs)
	login := authn.NewLogin(deps.AccountAuth, issuer)

	handler := transport.NewHandler(transport.Deps{
		Login:  login,
		Logger: deps.Logger,
	})

	return &Module{
		Issuer:  issuer,
		Login:   login,
		handler: handler,
	}
}

// Mount registers the authn module's HTTP routes on the given router.
func (m *Module) Mount(router gin.IRouter) {
	m.handler.RegisterRoutes(router)
}
