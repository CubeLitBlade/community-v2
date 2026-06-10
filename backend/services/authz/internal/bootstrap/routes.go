package bootstrap

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/health"
)

type registerParams struct {
	fx.In

	Router   *gin.Engine
	Mounters []platform.HTTPMounter `group:"mounter"`
	Logger   *slog.Logger
}

func registerAPIRoutes(p registerParams) {
	api := p.Router.Group("/api")
	platform.RegisterRouters(api, p.Mounters)
}

type healthRouteParams struct {
	fx.In

	Router        *gin.Engine
	HealthHandler *health.Handler
}

func registerHealthRoutes(p healthRouteParams) {
	p.HealthHandler.RegisterRoutes(p.Router)
}
