package redis

import (
	"context"
	"fmt"
	"slices"

	"github.com/redis/go-redis/v9"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

// Invalidator is an adapter that removes cached authorization decisions from Redis when underlying relations change.
//
// It relies on reverse indexes to locate all cache keys associated with the mutated users or objects.
type Invalidator struct {
	rdb *redis.Client
}

// NewInvalidator creates a new Redis cache invalidator.
func NewInvalidator(rdb *redis.Client) port.CacheInvalidator {
	return &Invalidator{
		rdb: rdb,
	}
}

// Invalidate deletes all cached authorization decisions associated with the given keys.
//
// It first queries the user and object index sets to find the exact cache keys to remove,
// then performs a bulk delete.
func (i *Invalidator) Invalidate(ctx context.Context, keys []string) error {
	cmds, err := i.getKeys(ctx, keys)
	if err != nil {
		return err
	}

	members, err := i.getMembers(cmds)
	if err != nil {
		return err
	}

	if members == nil {
		return nil
	}

	if err := i.rdb.Del(ctx, members...).Err(); err != nil {
		return fmt.Errorf("delete items: %w", err)
	}

	return nil
}

func (i *Invalidator) getKeys(ctx context.Context, keys []string) ([]*redis.StringSliceCmd, error) {
	copied := make([]string, len(keys))
	copy(copied, keys)

	slices.Sort(copied)
	copied = slices.Compact(copied)

	cmds := make([]*redis.StringSliceCmd, 0, len(keys)*2)

	_, err := i.rdb.Pipelined(ctx, func(p redis.Pipeliner) error {
		for _, key := range copied {
			cmds = append(cmds, p.SMembers(ctx, "authz:index:user:"+key))
			cmds = append(cmds, p.SMembers(ctx, "authz:index:object:"+key))
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("find indexes: %w", err)
	}

	return cmds, nil
}

func (*Invalidator) getMembers(cmds []*redis.StringSliceCmd) ([]string, error) {
	members := make([]string, 0, len(cmds))

	for _, cmd := range cmds {
		result, err := cmd.Result()
		if err != nil {
			return nil, fmt.Errorf("find members: %w", err)
		}

		members = append(members, result...)
	}

	if len(members) == 0 {
		return nil, nil
	}

	slices.Sort(members)
	members = slices.Compact(members)

	return members, nil
}
