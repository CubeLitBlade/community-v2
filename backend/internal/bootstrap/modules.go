package bootstrap

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	accountpostgres "github.com/CubeLitBlade/community-v2/backend/internal/account/postgres"
	accounttransport "github.com/CubeLitBlade/community-v2/backend/internal/account/transport"
	authtransport "github.com/CubeLitBlade/community-v2/backend/internal/auth/transport"
)

// ModuleDeps holds the shared dependencies required by application modules.
type ModuleDeps struct {
	DB     *gorm.DB
	IDs    account.IDGenerator
	Logger *slog.Logger
}

// RegisterModules registers all application modules and their routes onto the
// given router.
func RegisterModules(router *gin.Engine, deps ModuleDeps) {
	api := router.Group("/api")

	registerAccountModule(api, deps)
	registerAuthModule(api)
}

func registerAccountModule(router gin.IRouter, deps ModuleDeps) {
	accountReader := accountpostgres.NewReader(deps.DB)
	accountWriter := accountpostgres.NewWriter(deps.DB)
	accountService := account.NewService(
		deps.IDs,
		accountWriter,
		accountReader,
		deps.Logger,
	)

	accountHandler := accounttransport.NewHandler(accountService, deps.Logger)
	accountHandler.RegisterRoutes(router)
}

func registerAuthModule(router gin.IRouter) {
	authHandler := authtransport.NewHandler()
	authHandler.RegisterRoutes(router)
}
