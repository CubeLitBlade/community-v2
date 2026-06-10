// Package setup provides the fx module for the health check handler.
package setup

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/authz"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/health"
)

// Module returns the fx providers for the health module.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(NewPingerAdapter, fx.As(new(health.Pinger))),
		),
		fx.Provide(health.NewHandler),
	)
}

// NewPingerAdapter adapts authz.HealthChecker to satisfy health.Pinger.
func NewPingerAdapter(checker *authz.HealthChecker) *PingerAdapter {
	return &PingerAdapter{checker: checker}
}

// PingerAdapter adapts authz.Checker to satisfy health.Pinger.
type PingerAdapter struct {
	checker *authz.HealthChecker
}

// Ping delegates to authz.Checker.Ping.
func (a *PingerAdapter) Ping(ctx context.Context) error {
	if err := a.checker.Ping(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	return nil
}
