// Package snowflake provides a Snowflake ID generator constructor wired for fx.
package snowflake

import (
	"fmt"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/pkg/platform/config"
)

// Provide creates a Snowflake ID generator using the given config.
func Provide(cfg *config.SnowflakeConfig) (*idgen.Snowflake, error) {
	s, err := idgen.NewSnowflake(cfg.NodeID)
	if err != nil {
		return nil, fmt.Errorf("create snowflake: %w", err)
	}
	return s, nil
}
