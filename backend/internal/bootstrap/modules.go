package bootstrap

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	accountSetup "github.com/CubeLitBlade/community-v2/backend/internal/account/setup"
	authSetup "github.com/CubeLitBlade/community-v2/backend/internal/auth/setup"
	"github.com/CubeLitBlade/community-v2/backend/internal/authn"
	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
	"github.com/CubeLitBlade/community-v2/backend/internal/jwt"
)

const defaultAccessTokenTTLHours = 24

// ModuleDeps holds the shared dependencies required by application modules.
type ModuleDeps struct {
	DB     *gorm.DB
	IDs    idgen.Generator
	JWTKey string
	Logger *slog.Logger
}

// HTTPMounter is implemented by modules that register routes on a Gin router.
type HTTPMounter interface {
	Mount(router gin.IRouter)
}

// RegisterModules registers all application modules and their routes onto the
// given router.
func RegisterModules(router *gin.Engine, deps ModuleDeps) {
	// setup API group and middlewares
	api := router.Group("/api")
	authnMW := authn.Middleware(deps.JWTKey, deps.Logger)

	// initialize modules with dependencies
	accountMod := accountSetup.NewModule(accountSetup.ModuleDeps{
		DB:     deps.DB,
		IDs:    deps.IDs,
		Logger: deps.Logger,
	})

	authMod := authSetup.NewModule(authSetup.ModuleDeps{
		IDs: deps.IDs,
		JWTConfig: &jwt.Config{
			Key:      deps.JWTKey,
			Issuer:   "community-v2",
			Validity: time.Hour * defaultAccessTokenTTLHours,
		},
		AuthnMW:     authnMW,
		AccountAuth: accountMod.Authenticator,
		Logger:      deps.Logger,
	})

	// mount module routes
	mounters := []HTTPMounter{
		accountMod,
		authMod,
	}

	for _, mod := range mounters {
		mod.Mount(api)
	}
}
