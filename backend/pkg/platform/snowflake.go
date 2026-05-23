package platform

import (
	"fmt"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
)

// SnowflakeConfig holds the worker node ID for Snowflake ID generation.
type SnowflakeConfig struct {
	NodeID int64
}

// ProvideSnowflake creates a Snowflake ID generator using the given config.
func ProvideSnowflake(cfg *SnowflakeConfig) (*idgen.Snowflake, error) {
	s, err := idgen.NewSnowflake(cfg.NodeID)
	if err != nil {
		return nil, fmt.Errorf("create snowflake: %w", err)
	}
	return s, nil
}
