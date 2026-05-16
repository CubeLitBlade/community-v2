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
)

// ModuleDeps holds the external dependencies required to initialize the account
// module.
type ModuleDeps struct {
	DB     *gorm.DB
	IDs    account.IDGenerator
	Logger *slog.Logger
}

// RegisterAccountModule initializes the account module (including repository
// and domain services) and registers its HTTP handlers on the provided router.
func RegisterAccountModule(router gin.IRouter, deps ModuleDeps) {
	writer := postgres.NewWriter(deps.DB)

	registrar := account.NewRegistrar(deps.IDs, writer, deps.Logger)

	handler := transport.NewHandler(transport.Deps{
		Registrar: registrar,
		Logger:    deps.Logger,
	})
	handler.RegisterRoutes(router)
}
