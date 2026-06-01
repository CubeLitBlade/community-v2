// Package setup provides constructors that wire together auth domain services,
// JWT issuance, and HTTP transport.
package setup

import (
	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/auth"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/auth/transport"
	"go.uber.org/fx"
)

// Module returns the fx providers for the auth module.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(auth.NewLogin),
		fx.Provide(
			fx.Annotate(transport.NewHandler, fx.As(new(platform.HTTPMounter)), fx.ResultTags(`group:"mounter"`)),
		),
	)
}
