package bootstrap

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// HTTPMounter is implemented by HTTP handlers that register routes on a Gin router.
type HTTPMounter interface {
	RegisterRoutes(router gin.IRouter)
}

type routeMountParams struct {
	fx.In

	Router   *gin.Engine
	Mounters []HTTPMounter `group:"mounter"`
}

func registerRoutes(p routeMountParams) {
	api := p.Router.Group("/api")

	for _, m := range p.Mounters {
		m.RegisterRoutes(api)
	}
}
