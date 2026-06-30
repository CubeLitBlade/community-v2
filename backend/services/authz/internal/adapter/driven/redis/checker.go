package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

// Checker is an adapter that reads cached authorization decisions from Redis.
type Checker struct {
	rdb *redis.Client
}

// NewChecker creates a new Redis cache checker.
func NewChecker(rdb *redis.Client) port.CacheChecker {
	return &Checker{
		rdb: rdb,
	}
}

// Check retrieves the cached authorization decision. It returns port.ErrCacheMiss if the key does not exist.
func (c *Checker) Check(ctx context.Context, user, relation, object string) (bool, error) {
	key := fmt.Sprintf("authz:check:%s:%s:%s", user, relation, object)

	allowed, err := c.rdb.Get(ctx, key).Bool()
	switch {
	case err == nil:
		return allowed, nil
	case errors.Is(err, redis.Nil):
		return false, port.ErrCacheMiss
	default:
		return false, fmt.Errorf("redis error: %w", err)
	}
}
