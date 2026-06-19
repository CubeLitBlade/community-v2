package bootstrap

import (
	"fmt"

	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/shared"
)

func provideSnowflake(cfg *shared.SnowflakeConfig) (*idgen.Snowflake, error) {
	s, err := idgen.NewSnowflake(cfg.NodeID)
	if err != nil {
		return nil, fmt.Errorf("create snowflake: %w", err)
	}

	return s, nil
}

// SnowflakeModule provides the Snowflake ID generator fx module.
func SnowflakeModule() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(provideSnowflake, fx.As(new(idgen.Generator))),
		),
	)
}
