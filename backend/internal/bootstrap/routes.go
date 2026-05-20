package bootstrap

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/CubeLitBlade/community-v2/backend/internal/authn"
)

// HTTPMounter is implemented by HTTP handlers that register routes on a Gin router.
type HTTPMounter interface {
	RegisterRoutes(router gin.IRouter)
}

type routeMountParams struct {
	fx.In

	cfg    Config
	logger *slog.Logger

	Router   *gin.Engine
	Mounters []HTTPMounter `group:"mounter"`
}

func registerRoutes(p routeMountParams) {
	api := p.Router.Group("/api")
	api.Use(authn.ParseToken(p.cfg.JWTSecret, p.logger))

	for _, m := range p.Mounters {
		m.RegisterRoutes(api)
	}
}
