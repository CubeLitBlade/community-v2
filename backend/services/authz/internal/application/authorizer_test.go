package application_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/application"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

var (
	_ port.AuthorityChecker = (*FakeAuthChecker)(nil)
	_ port.CacheChecker     = (*FakeCacheChecker)(nil)
	_ port.CacheWriter      = (*FakeCacheWriter)(nil)
)

var errFake = errors.New("fake error")

const (
	UserAlice          = "alice"
	AdminBob           = "bob"
	RelationEditPost   = "can_edit_post"
	RelationDeletePost = "can_delete_post"
	PostAlice          = "post_belongs_to_alice"
	PostBob            = "post_belongs_to_bob"
)

type FakeCacheChecker struct {
	CalledCount int
	errToReturn error
	data        map[string]bool
}

func NewFakeCacheChecker(data map[string]bool) *FakeCacheChecker {
	return &FakeCacheChecker{data: data}
}

func (c *FakeCacheChecker) Check(_ context.Context, user, relation, object string) (bool, error) {
	c.CalledCount++

	if c.errToReturn != nil {
		return false, c.errToReturn
	}

	allowed, ok := c.data[key(user, relation, object)]
	if !ok {
		return false, port.ErrCacheMiss
	}

	return allowed, nil
}

type FakeCacheWriter struct {
	CalledCount int
	errToReturn error
	data        map[string]bool
}

func NewFakeCacheWriter(data map[string]bool) *FakeCacheWriter {
	return &FakeCacheWriter{data: data}
}

func (w *FakeCacheWriter) Write(_ context.Context, user, relation, object string, allowed bool, _ time.Duration) error {
	w.CalledCount++

	if w.errToReturn != nil {
		return w.errToReturn
	}

	w.data[key(user, relation, object)] = allowed

	return nil
}

type FakeAuthChecker struct {
	CalledCount int
	errToReturn error
	data        map[string]bool
}

func NewFakeAuthChecker(data map[string]bool) *FakeAuthChecker {
	return &FakeAuthChecker{data: data}
}

func (c *FakeAuthChecker) Check(_ context.Context, user, relation, object string) (bool, error) {
	c.CalledCount++

	if c.errToReturn != nil {
		return false, c.errToReturn
	}

	allowed, ok := c.data[key(user, relation, object)]
	if !ok {
		return false, nil
	}

	return allowed, nil
}

func key(user, relation, object string) string {
	return fmt.Sprintf("%s|%s|%s", user, relation, object)
}

func cloneMap(src map[string]bool) map[string]bool {
	dst := make(map[string]bool, len(src))
	maps.Copy(dst, src)

	return dst
}

//nolint:revive // for testing only
func TestAuthorizer_Authorize(t *testing.T) {
	t.Parallel()

	// Template maps — each test case clones these to avoid data races under t.Parallel().
	cacheTemplate := map[string]bool{
		key(UserAlice, RelationEditPost, PostAlice): true,
	}
	authTemplate := map[string]bool{
		key(UserAlice, RelationEditPost, PostAlice):  true,
		key(AdminBob, RelationDeletePost, PostAlice): true,
		key(AdminBob, RelationEditPost, PostBob):     true,
		key(UserAlice, RelationDeletePost, PostBob):  false,
	}

	tests := []struct {
		name string
		// Authorize inputs.
		user     string
		relation string
		object   string
		// Expected outputs.
		want    bool
		wantErr error
		// Expected call counts.
		wantAuthCheckerCalls  int
		wantCacheCheckerCalls int
		wantCacheWriterCalls  int
		// Optional: force an error on a specific fake.
		cacheCheckerErr error
		authCheckerErr  error
		cacheWriterErr  error
		// Optional: skip cache data so cache checker always returns ErrCacheMiss.
		skipCacheData bool
		// Optional: skip auth data so auth checker always returns false,nil.
		skipAuthData bool
	}{
		{
			name:                  "cache hit: return directly",
			user:                  UserAlice,
			relation:              RelationEditPost,
			object:                PostAlice,
			want:                  true,
			wantAuthCheckerCalls:  0,
			wantCacheCheckerCalls: 1,
			wantCacheWriterCalls:  0,
		},
		{
			name:                  "cache miss, auth allows, cache written",
			user:                  AdminBob,
			relation:              RelationDeletePost,
			object:                PostAlice,
			want:                  true,
			wantAuthCheckerCalls:  1,
			wantCacheCheckerCalls: 1,
			wantCacheWriterCalls:  1,
			skipCacheData:         true,
		},
		{
			name:                  "cache miss, auth denies, cache written",
			user:                  UserAlice,
			relation:              RelationDeletePost,
			object:                PostBob,
			want:                  false,
			wantAuthCheckerCalls:  1,
			wantCacheCheckerCalls: 1,
			wantCacheWriterCalls:  1,
			skipCacheData:         true,
		},
		{
			name:                  "cache miss, auth check error",
			user:                  AdminBob,
			relation:              RelationDeletePost,
			object:                PostAlice,
			want:                  false,
			wantErr:               errFake,
			wantAuthCheckerCalls:  1,
			wantCacheCheckerCalls: 1,
			wantCacheWriterCalls:  0,
			skipCacheData:         true,
			authCheckerErr:        errFake,
		},
		{
			name:                  "cache checker unexpected error, fallback to auth",
			user:                  AdminBob,
			relation:              RelationEditPost,
			object:                PostBob,
			want:                  true,
			wantAuthCheckerCalls:  1,
			wantCacheCheckerCalls: 1,
			wantCacheWriterCalls:  1,
			cacheCheckerErr:       errFake,
		},
		{
			name:                  "cache miss, auth allows, cache write fails but returns success",
			user:                  AdminBob,
			relation:              RelationDeletePost,
			object:                PostAlice,
			want:                  true,
			wantErr:               nil,
			wantAuthCheckerCalls:  1,
			wantCacheCheckerCalls: 1,
			wantCacheWriterCalls:  1,
			skipCacheData:         true,
			cacheWriterErr:        errFake,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Each test case gets its own copy of the data maps.
			cacheData := cloneMap(cacheTemplate)
			if tt.skipCacheData {
				cacheData = make(map[string]bool)
			}

			authData := cloneMap(authTemplate)
			if tt.skipAuthData {
				authData = make(map[string]bool)
			}

			cacheChecker := NewFakeCacheChecker(cacheData)
			cacheChecker.errToReturn = tt.cacheCheckerErr

			cacheWriter := NewFakeCacheWriter(cacheData)
			cacheWriter.errToReturn = tt.cacheWriterErr

			authChecker := NewFakeAuthChecker(authData)
			authChecker.errToReturn = tt.authCheckerErr

			a := application.NewAuthorizer(
				authChecker, cacheChecker, cacheWriter, slog.Default(), noop.NewTracerProvider(),
			)

			got, gotErr := a.Authorize(context.Background(), tt.user, tt.relation, tt.object)

			if tt.wantErr != nil {
				assert.ErrorIs(t, gotErr, tt.wantErr)
				assert.Equal(t, tt.wantAuthCheckerCalls, authChecker.CalledCount)
				assert.Equal(t, tt.wantCacheCheckerCalls, cacheChecker.CalledCount)
				assert.Equal(t, tt.wantCacheWriterCalls, cacheWriter.CalledCount)

				return
			}

			assert.NoError(t, gotErr)
			assert.Equal(t, tt.want, got)

			assert.Equal(t, tt.wantAuthCheckerCalls, authChecker.CalledCount)
			assert.Equal(t, tt.wantCacheCheckerCalls, cacheChecker.CalledCount)
			assert.Equal(t, tt.wantCacheWriterCalls, cacheWriter.CalledCount)
		})
	}
}
