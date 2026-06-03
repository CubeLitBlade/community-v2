package bootstrap

import (
	"fmt"

	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func ginMode(appRT AppRuntime) string {
	if appRT.Environment() == AppEnvProduction {
		return gin.ReleaseMode
	}

	return gin.DebugMode
}

func provideGinEngine(appRT AppRuntime) *gin.Engine {
	gin.SetMode(ginMode(appRT))

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
