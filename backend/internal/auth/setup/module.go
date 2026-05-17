package setup

import (
	"log/slog"

	"github.com/CubeLitBlade/community-v2/backend/internal/auth"
	"github.com/CubeLitBlade/community-v2/backend/internal/auth/transport"
	"github.com/gin-gonic/gin"
)

type Module struct {
	Issuer *auth.JWTIssuer
	Login  *auth.Login

	handler *transport.Handler
}

type ModuleDeps struct {
	IDs         auth.IDGenerator
	JWTConfig   *auth.JWTConfig
	AccountAuth auth.AccountAuthenticator
	Logger      *slog.Logger
}

func NewModule(deps ModuleDeps) *Module {
	issuer := auth.NewJWTIssuer(deps.JWTConfig, deps.IDs)
	login := auth.NewLogin(deps.AccountAuth, issuer)

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

func (m *Module) Mount(router gin.IRouter) {
	m.handler.RegisterRoutes(router)
}
