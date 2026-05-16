package bootstrap

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/CubeLitBlade/community-v2/backend/internal/account/setup"
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

	setup.RegisterAccountModule(api, setup.ModuleDeps{
		DB:     deps.DB,
		IDs:    deps.IDs,
		Logger: deps.Logger,
	})
	registerAuthModule(api)
}

func registerAuthModule(router gin.IRouter) {
	authHandler := authtransport.NewHandler()
	authHandler.RegisterRoutes(router)
}
