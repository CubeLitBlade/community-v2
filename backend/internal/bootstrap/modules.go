package bootstrap

import (
	"log/slog"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/CubeLitBlade/community-v2/backend/internal/web"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	accountpostgres "github.com/CubeLitBlade/community-v2/backend/internal/account/postgres"
)

type ModuleDeps struct {
	DB     *gorm.DB
	IDs    account.IDGenerator
	Logger *slog.Logger
}

func RegisterModules(router *gin.Engine, deps ModuleDeps) {
	registerAccountModule(router, deps)
}

func registerAccountModule(router *gin.Engine, deps ModuleDeps) {
	accountRepo := accountpostgres.NewRepository(deps.DB)
	accountService := account.NewService(deps.IDs, accountRepo, deps.Logger)

	accountHandler := web.NewAccountHandler(accountService, deps.Logger)
	accountHandler.RegisterRoutes(router)
}
