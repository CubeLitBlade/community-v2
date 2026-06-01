package bootstrap

import (
	"log/slog"

	"github.com/cubelitblade/community-v2/backend/pkg/common/jwt"
	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/auth/transport"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/authn"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/health"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type registerParams struct {
	fx.In

	Router           *gin.Engine
	Mounters         []platform.HTTPMounter `group:"mounter"`
	Parser           *jwt.Parser
	AuthCookieConfig *transport.CookieConfig
	Logger           *slog.Logger
}

func registerAPIRoutes(p registerParams) {
	api := p.Router.Group("/api")
	api.Use(authn.ParseToken(p.Parser, p.AuthCookieConfig.Name, p.Logger))
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
