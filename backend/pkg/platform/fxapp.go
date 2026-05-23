package platform

import (
	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"go.uber.org/fx"
)

// StandardModules returns the standard set of infrastructure providers
// (logger, database, Gin engine, HTTP server, Snowflake ID generator).
func StandardModules() fx.Option {
	return fx.Options(
		fx.Provide(NewLogger),
		fx.Provide(NewGormLogger),
		fx.Provide(ProvideDatabase),
		fx.Provide(NewGinEngine),
		fx.Provide(
			fx.Annotate(ProvideSnowflake, fx.As(new(idgen.Generator))),
		),
		fx.Provide(ProvideHTTPServer),
	)
}
