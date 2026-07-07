// Package setup provides the fx module for the health check handler.
package setup

import (
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/health"
)

// Module returns the fx providers for the health module.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(health.NewHandler),
	)
}
