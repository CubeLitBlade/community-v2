package bootstrap

import (
	"log/slog"

	"github.com/cubelitblade/community-v2/backend/pkg/common/jwt"
	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	v1 "github.com/cubelitblade/community-v2/backend/services/account/api/rest/v1"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/authn"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/health"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type registerParams struct {
	fx.In

	Router   *gin.Engine
	Mounters []platform.HTTPMounter `group:"mounter"`
	Parser   *jwt.Parser
	Logger   *slog.Logger
}

func registerAPIRoutes(p registerParams) {
	api := p.Router.Group("/api")
	api.Use(authn.ParseToken(p.Parser, v1.AccessTokenCookieName, p.Logger))
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
