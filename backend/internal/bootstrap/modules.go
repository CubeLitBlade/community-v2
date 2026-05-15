package bootstrap

import (
	"log/slog"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	accountpostgres "github.com/CubeLitBlade/community-v2/backend/internal/account/postgres"   //nolint:revive // import path cannot be shortened
	accounttransport "github.com/CubeLitBlade/community-v2/backend/internal/account/transport" //nolint:revive // import path cannot be shortened
	authtransport "github.com/CubeLitBlade/community-v2/backend/internal/auth/transport"       //nolint:revive // import path cannot be shortened
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	accountRepo := accountpostgres.NewRepository(deps.DB)
	accountService := account.NewService(deps.IDs, accountRepo, deps.Logger)

	accountHandler := accounttransport.NewHandler(accountService, deps.Logger)
	accountHandler.RegisterRoutes(router)
}

func registerAuthModule(router gin.IRouter) {
	authHandler := authtransport.NewHandler()
	authHandler.RegisterRoutes(router)
}
