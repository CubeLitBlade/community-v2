package bootstrap

import (
	"log/slog"
	"time"

	"github.com/CubeLitBlade/community-v2/backend/internal/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	accountSetup "github.com/CubeLitBlade/community-v2/backend/internal/account/setup"
	authSetup "github.com/CubeLitBlade/community-v2/backend/internal/auth/setup"
)

// ModuleDeps holds the shared dependencies required by application modules.
type ModuleDeps struct {
	DB     *gorm.DB
	IDs    account.IDGenerator
	JWTKey string
	Logger *slog.Logger
}

type HTTPMounter interface {
	Mount(router gin.IRouter)
}

// RegisterModules registers all application modules and their routes onto the
// given router.
func RegisterModules(router *gin.Engine, deps ModuleDeps) {
	api := router.Group("/api")
	accountMod := accountSetup.NewModule(accountSetup.ModuleDeps{
		DB:     deps.DB,
		IDs:    deps.IDs,
		Logger: deps.Logger,
	})

	authMod := authSetup.NewModule(authSetup.ModuleDeps{
		IDs: deps.IDs,
		JWTConfig: &auth.JWTConfig{
			Key:      deps.JWTKey,
			Issuer:   "community-v2",
			Validity: time.Hour * 24,
		},
		AccountAuth: accountMod.Authenticator,
		Logger:      deps.Logger,
	})

	// mount APIs
	mounters := []HTTPMounter{
		accountMod,
		authMod,
	}

	for _, mod := range mounters {
		mod.Mount(api)
	}
}
