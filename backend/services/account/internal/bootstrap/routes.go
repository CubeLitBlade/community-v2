package bootstrap

import (
	"log/slog"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/authn"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// HTTPMounter is implemented by HTTP handlers that register routes on a Gin router.
type HTTPMounter interface {
	RegisterRoutes(router gin.IRouter)
}

type routeMountParams struct {
	fx.In

	Config *Config
	Logger *slog.Logger

	Router   *gin.Engine
	Mounters []HTTPMounter `group:"mounter"`
}

func registerRoutes(p routeMountParams) {
	api := p.Router.Group("/api")
	api.Use(authn.ParseToken(p.Config.JWTSecret, p.Config.AccessTokenCookieName, p.Logger))

	for _, m := range p.Mounters {
		m.RegisterRoutes(api)
	}
}
