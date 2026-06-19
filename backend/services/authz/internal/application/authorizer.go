package application

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

type Authorizer struct {
	checker port.Checker
	logger  *slog.Logger
}

func NewAuthorizer(checker port.Checker, logger *slog.Logger) *Authorizer {
	return &Authorizer{
		checker: checker,
		logger:  logger.With(slog.String("component", "authorizer")),
	}
}

func (c *Authorizer) Authorize(ctx context.Context, user, relation, object string) (bool, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("authz.user", user),
		attribute.String("authz.relation", relation),
		attribute.String("authz.object", object),
	)

	allowed, err := c.checker.Check(ctx, user, relation, object)
	if err != nil {
		c.logger.ErrorContext(ctx, "authorize call went sideways",
			slog.String("user", user),
			slog.String("relation", relation),
			slog.String("object", object),
			slog.Any("error", err),
		)

		span.SetStatus(codes.Error, "failed to authorize")
		span.RecordError(err)

		return false, fmt.Errorf("authorize: %w", err)
	}

	c.logger.DebugContext(ctx, "authorization pronounced",
		slog.String("user", user),
		slog.String("relation", relation),
		slog.String("object", object),
		slog.Bool("allowed", allowed),
	)
	span.SetAttributes(attribute.Bool("authz.allowed", allowed))

	return allowed, nil
}
