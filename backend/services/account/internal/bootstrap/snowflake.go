package bootstrap

import (
	"fmt"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"go.uber.org/fx"
)

// SnowflakeConfig holds the Snowflake ID generator configuration.
type SnowflakeConfig struct {
	SnowflakeID int64 `env:"SNOWFLAKE_ID" required:"true"`
}

func provideSnowflake(cfg SnowflakeConfig) (*idgen.Snowflake, error) {
	s, err := idgen.NewSnowflake(cfg.SnowflakeID)
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
