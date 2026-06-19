package bootstrap

import (
	"fmt"

	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/idgenerator"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/config"
)

func NewSnowflake(cfg config.SnowflakeConfig) (*idgen.Snowflake, error) {
	ids, err := idgen.NewSnowflake(cfg.NodeID)
	if err != nil {
		return nil, fmt.Errorf("generate snowflake instance: %w", err)
	}

	return ids, nil
}

func SnowflakeModule() fx.Option {
	return fx.Options(
		fx.Provide(NewSnowflake),
		fx.Provide(idgenerator.New),
	)
}
