package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

const (
	decisionCacheExpiration = 5 * time.Minute
)

// AuthChecker evaluates whether a user is granted a relation on an object.
type AuthChecker interface {
	Authorize(ctx context.Context, user, relation, object string) (bool, error)
}

// Authorizer evaluates user permissions using a cache and an authority checker.
type Authorizer struct {
	authChecker  port.AuthorityChecker
	cacheChecker port.CacheChecker
	cacheWriter  port.CacheWriter
	logger       *slog.Logger
}

// NewAuthorizer creates a new Authorizer instance.
func NewAuthorizer(
	authChecker port.AuthorityChecker,
	cacheChecker port.CacheChecker,
	cacheWriter port.CacheWriter,
	logger *slog.Logger,
) *Authorizer {
	return &Authorizer{
		authChecker:  authChecker,
		cacheChecker: cacheChecker,
		cacheWriter:  cacheWriter,
		logger:       logger.With(slog.String("component", "authorizer")),
	}
}

// Authorize determines if a user is granted a relation on an object.
//
// It checks the cache first, falling back to the AuthorityChecker on a miss,
// and caches the result upon a successful evaluation.
func (a *Authorizer) Authorize(ctx context.Context, user, relation, object string) (bool, error) {
	allowed, cerr := a.cacheChecker.Check(ctx, user, relation, object)
	switch {
	case cerr == nil:
		a.logger.DebugContext(ctx, "cache hit")

		return allowed, nil
	case errors.Is(cerr, port.ErrCacheMiss):
		a.logger.DebugContext(ctx, "cache miss")
	default:
		a.logger.ErrorContext(ctx, "unexpected cache error", slog.Any("error", cerr))
	}

	allowed, serr := a.authChecker.Check(ctx, user, relation, object)
	if serr != nil {
		a.logger.ErrorContext(ctx, "authorize call went sideways",
			slog.String("user", user),
			slog.String("relation", relation),
			slog.String("object", object),
			slog.Any("error", serr),
		)

		return false, fmt.Errorf("authorize: %w", serr)
	}

	if werr := a.cacheWriter.Write(ctx, user, relation, object, allowed, decisionCacheExpiration); werr != nil {
		a.logger.ErrorContext(ctx, "failed to write cache", slog.Any("error", werr))
	}

	a.logger.DebugContext(ctx, "authorization pronounced",
		slog.String("user", user),
		slog.String("relation", relation),
		slog.String("object", object),
		slog.Bool("allowed", allowed),
	)

	return allowed, nil
}
