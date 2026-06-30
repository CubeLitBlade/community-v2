package redis

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

const (
	// primaryTTLJitterSecs defines the maximum random jitter added to the primary cache key's TTL to
	// prevent cache stampedes (thundering herd) when many cache entries expire at the exact same time.
	primaryTTLJitterSecs = 60

	// indexTTLJitterSecs defines the maximum random jitter added to the index sets' TTLs.
	// The index TTL is based on the primary TTL to ensure indexes outlive their associated cache keys,
	// allowing the Invalidator to clean them up.
	indexTTLJitterSecs = 10
)

var errInvalidExpiration = errors.New("expiration must be greater than 0")

// Writer is an adapter that saves authorization decisions to Redis.
//
// It writes the primary cache entry and updates the corresponding reverse indexes
// atomically to support future cache invalidation.
type Writer struct {
	rdb *redis.Client
}

// NewWriter creates a new Redis cache writer.
func NewWriter(rdb *redis.Client) port.CacheWriter {
	return &Writer{
		rdb: rdb,
	}
}

// Write stores an authorization decision in Redis and updates the associated user and object reverse indexes.
//
// It uses a transaction pipeline to ensure both the cache key and indexes are written atomically.
func (w *Writer) Write(
	ctx context.Context, user, relation, object string, allowed bool, expiration time.Duration,
) error {
	if expiration <= 0 {
		return errInvalidExpiration
	}

	key := fmt.Sprintf("authz:check:%s:%s:%s", user, relation, object)
	userIndex := "authz:index:user:" + user
	objectIndex := "authz:index:object:" + object

	expiration += time.Duration(rand.IntN(primaryTTLJitterSecs)) * time.Second

	_, err := w.rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
		p.Set(ctx, key, allowed, expiration)

		// add the primary key to reverse indexes for future invalidation lookups
		p.SAdd(ctx, userIndex, key)
		p.SAdd(ctx, objectIndex, key)

		// set TTL for reverse indexes
		// the index TTL is based on the primary TTL (which already includes jitter) plus an additional small jitter
		// this ensures indexes outlive the primary keys preventing premature expiration during invalidation
		p.Expire(ctx, userIndex, expiration+time.Duration(rand.IntN(indexTTLJitterSecs))*time.Second)
		p.Expire(ctx, objectIndex, expiration+time.Duration(rand.IntN(indexTTLJitterSecs))*time.Second)

		return nil
	})
	if err != nil {
		return fmt.Errorf("tx pipeline: %w", err)
	}

	return nil
}
