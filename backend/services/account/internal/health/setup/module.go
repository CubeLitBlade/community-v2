// Package setup provides the fx module for the health check handler.
package setup

import (
	"github.com/cubelitblade/community-v2/backend/services/account/internal/health"
	"go.uber.org/fx"
)

// Module returns the fx providers for the health module.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(health.NewHandler),
	)
}
