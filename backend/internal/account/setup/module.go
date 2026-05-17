// Package setup wires together the dependencies for the account module
// and registers its HTTP routes.
package setup

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/CubeLitBlade/community-v2/backend/internal/account/postgres"
	"github.com/CubeLitBlade/community-v2/backend/internal/account/transport"
	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
)

// ModuleDeps holds the external dependencies required to initialize the account
// module.
type ModuleDeps struct {
	DB     *gorm.DB
	IDs    idgen.Generator
	Logger *slog.Logger
}

// Module is the account domain module, holding its services and HTTP handler.
type Module struct {
	Registrar     *account.Registrar
	Authenticator *account.Authenticator

	handler *transport.Handler
}

// NewModule wires the account domain services and returns a Module.
func NewModule(deps ModuleDeps) *Module {
	writer := postgres.NewWriter(deps.DB)
	reader := postgres.NewReader(deps.DB)

	registrar := account.NewRegistrar(deps.IDs, writer, deps.Logger)
	authenticator := account.NewAuthenticator(reader, deps.Logger)

	handler := transport.NewHandler(transport.Deps{
		Registrar: registrar,
		Logger:    deps.Logger,
	})

	return &Module{
		Registrar:     registrar,
		Authenticator: authenticator,
		handler:       handler,
	}
}

// Mount initializes the account module (including repository
// and domain services) and registers its HTTP handlers on the provided router.
func (m *Module) Mount(router gin.IRouter) {
	m.handler.RegisterRoutes(router)
}
