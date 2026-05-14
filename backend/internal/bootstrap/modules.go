package bootstrap

import (
	"log/slog"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	accountpostgres "github.com/CubeLitBlade/community-v2/backend/internal/account/postgres"
	"github.com/CubeLitBlade/community-v2/backend/internal/web"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ModuleDeps struct {
	DB     *gorm.DB
	IDs    account.IDGenerator
	Logger *slog.Logger
}

func RegisterModules(router *gin.Engine, deps ModuleDeps) {
	api := router.Group("/api")

	registerAccountModule(api, deps)
	registerAuthModule(api)
}

func registerAccountModule(router gin.IRouter, deps ModuleDeps) {
	accountRepo := accountpostgres.NewRepository(deps.DB)
	accountService := account.NewService(deps.IDs, accountRepo, deps.Logger)

	accountHandler := web.NewAccountHandler(accountService, deps.Logger)
	accountHandler.RegisterRoutes(router)
}

func registerAuthModule(router gin.IRouter) {
	authHandler := web.NewAuthHandler()
	authHandler.RegisterRoutes(router)
}
