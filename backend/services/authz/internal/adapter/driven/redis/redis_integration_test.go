package redis_test

import (
	"context"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	redisadapter "github.com/cubelitblade/community-v2/backend/services/authz/internal/adapter/driven/redis"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

func client(ctx context.Context, t *testing.T) *goredis.Client {
	t.Helper()

	redisContainer, err := tcredis.Run(ctx, "redis:8.8.0-alpine")
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	})

	uri, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	opt, err := goredis.ParseURL(uri)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	rdb := goredis.NewClient(opt)

	t.Cleanup(func() {
		rdb.Close()
	})

	return rdb
}

//nolint:paralleltest,revive // Intentionally sequential: subtests share a single Redis Testcontainer.
func Test_IntegrationRedisAdapter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping redis adapter integration test")
	}

	ctx := context.Background()
	rdb := client(ctx, t)

	checker := redisadapter.NewChecker(rdb)
	writer := redisadapter.NewWriter(rdb)
	invalidator := redisadapter.NewInvalidator(rdb)

	t.Run("write and check cache hit", func(t *testing.T) {
		const user, relation, object = "alice", "can_edit", "post_1"

		err := writer.Write(ctx, user, relation, object, true, time.Hour)
		require.NoError(t, err)

		allowed, err := checker.Check(ctx, user, relation, object)
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("write denied and check", func(t *testing.T) {
		const user, relation, object = "bob", "can_delete", "post_2"

		err := writer.Write(ctx, user, relation, object, false, time.Hour)
		require.NoError(t, err)

		allowed, err := checker.Check(ctx, user, relation, object)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("check without write returns cache miss", func(t *testing.T) {
		const user, relation, object = "charlie", "can_edit", "post_3"

		allowed, err := checker.Check(ctx, user, relation, object)
		assert.ErrorIs(t, err, port.ErrCacheMiss)
		assert.False(t, allowed)
	})

	t.Run("invalidate by user key", func(t *testing.T) {
		const user, relation, object = "dave", "can_edit", "post_4"

		err := writer.Write(ctx, user, relation, object, true, time.Hour)
		require.NoError(t, err)

		err = invalidator.Invalidate(ctx, []string{user})
		require.NoError(t, err)

		allowed, err := checker.Check(ctx, user, relation, object)
		assert.ErrorIs(t, err, port.ErrCacheMiss)
		assert.False(t, allowed)
	})

	t.Run("invalidate by object key", func(t *testing.T) {
		const user, relation, object = "eve", "can_edit", "post_5"

		err := writer.Write(ctx, user, relation, object, true, time.Hour)
		require.NoError(t, err)

		err = invalidator.Invalidate(ctx, []string{object})
		require.NoError(t, err)

		allowed, err := checker.Check(ctx, user, relation, object)
		assert.ErrorIs(t, err, port.ErrCacheMiss)
		assert.False(t, allowed)
	})

	t.Run("invalidate only removes targeted keys", func(t *testing.T) {
		const (
			user               = "frank"
			relationA, objectA = "can_edit", "post_6"
			relationB, objectB = "can_delete", "post_7"
		)

		// Write two tuples under the same user.
		err := writer.Write(ctx, user, relationA, objectA, true, time.Hour)
		require.NoError(t, err)
		err = writer.Write(ctx, user, relationB, objectB, false, time.Hour)
		require.NoError(t, err)

		// Invalidate only objectA.
		err = invalidator.Invalidate(ctx, []string{objectA})
		require.NoError(t, err)

		// objectA tuple should be evicted.
		allowed, err := checker.Check(ctx, user, relationA, objectA)
		assert.ErrorIs(t, err, port.ErrCacheMiss)
		assert.False(t, allowed)

		// objectB tuple should still be cached.
		allowed, err = checker.Check(ctx, user, relationB, objectB)
		require.NoError(t, err)
		assert.False(t, allowed)
	})
}
