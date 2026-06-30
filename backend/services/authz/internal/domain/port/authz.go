// Package port defines the secondary ports (driven-side interfaces) for the authz service.
package port

import (
	"context"
	"errors"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain"
)

// ErrCacheMiss indicates that a requested cache entry was not found.
var ErrCacheMiss = errors.New("cache miss")

// AuthorityChecker checks if a user has a specific relation to an object.
type AuthorityChecker interface {
	// Check returns true if the user is granted the relation on the object.
	Check(ctx context.Context, user, relation, object string) (bool, error)
}

// TupleWriter persists authorization tuples to storage.
type TupleWriter interface {
	// Write saves a batch of tuples to the underlying storage.
	Write(ctx context.Context, tuples []domain.Tuple) error
}

// CacheChecker looks up cached authorization decisions.
type CacheChecker interface {
	// Check returns the cached decision. It returns ErrCacheMiss if the entry is not found.
	Check(ctx context.Context, user, relation, object string) (bool, error)
}

// CacheWriter saves an authorization decision to the cache.
type CacheWriter interface {
	// Write stores the authorization decision with a specified expiration time.
	Write(ctx context.Context, user, relation, object string, allowed bool, expiration time.Duration) error
}

// CacheInvalidator evicts specific entries from the cache.
type CacheInvalidator interface {
	// Invalidate removes the specified keys from the cache.
	Invalidate(ctx context.Context, keys []string) error
}
