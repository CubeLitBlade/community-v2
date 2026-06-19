package bootstrap

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/shared"
)

func provideGinEngine(cfg *shared.GinConfig) *gin.Engine {
	gin.SetMode(strings.ToLower(cfg.Mode))

	handlers := []gin.HandlerFunc{
		gin.Logger(),
		gin.Recovery(),
		platform.RequestIDMiddleware(),
	}

	engine := platform.NewGinEngine(handlers...)

	if err := engine.SetTrustedProxies(nil); err != nil {
		// This shouldn't happen with a valid config; panic to fail fast
		panic(fmt.Sprintf("set trusted proxies: %v", err))
	}

	return engine
}

// GinModule provides the Gin HTTP engine fx module.
func GinModule() fx.Option {
	return fx.Options(
		fx.Provide(provideGinEngine),
	)
}
