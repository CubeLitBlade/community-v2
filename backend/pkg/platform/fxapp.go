package platform

import (
	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/pkg/platform/database"
	"github.com/cubelitblade/community-v2/backend/pkg/platform/logger"
	"github.com/cubelitblade/community-v2/backend/pkg/platform/server"
	"github.com/cubelitblade/community-v2/backend/pkg/platform/snowflake"
	"go.uber.org/fx"
)

// StandardModules returns the standard set of infrastructure providers
// (logger, database, Gin engine, HTTP server, Snowflake ID generator).
func StandardModules() fx.Option {
	return fx.Options(
		fx.Provide(logger.NewLogger),
		fx.Provide(logger.NewGormLogger),
		fx.Provide(database.Provide),
		fx.Provide(server.NewGinEngine),
		fx.Provide(
			fx.Annotate(snowflake.Provide, fx.As(new(idgen.Generator))),
		),
		fx.Provide(server.ProvideHTTPServer),
	)
}
