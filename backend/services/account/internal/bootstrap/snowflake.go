package bootstrap

import (
	"fmt"

	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/config"
)

func provideSnowflake(cfg config.SnowflakeConfig) (idgen.Generator, error) {
	s, err := idgen.NewSnowflake(cfg.NodeID)
	if err != nil {
		return nil, fmt.Errorf("create snowflake: %w", err)
	}

	return s, nil
}

func SnowflakeModule() fx.Option {
	return fx.Options(
		fx.Provide(provideSnowflake),
	)
}
