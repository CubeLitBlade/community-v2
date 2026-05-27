// Package setup provides constructors that wire together account domain services,
// persistence, and HTTP transport.
package setup

import (
	"context"
	"fmt"

	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/account"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/account/postgres"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/account/transport"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/auth"
	"go.uber.org/fx"
)

// Module returns the fx providers for the account module.
// Cross-module interface bindings (AccountAuthenticator, LoginRecorder)
// are handled by the composition root in bootstrap.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(postgres.NewWriter, fx.As(new(account.Creator))),
			fx.Annotate(postgres.NewWriter, fx.As(new(account.LastLoginUpdater))),
		),
		fx.Provide(
			fx.Annotate(postgres.NewReader, fx.As(new(account.ByIDFinder))),
			fx.Annotate(postgres.NewReader, fx.As(new(account.ByUsernameFinder))),
		),
		fx.Provide(
			fx.Annotate(account.NewRegistrar, fx.As(new(transport.Registrar))),
		),
		fx.Provide(account.NewAuthenticator),
		fx.Provide(
			fx.Annotate(account.NewProfileFinder, fx.As(new(transport.ProfileFinder))),
		),
		fx.Provide(NewAuthenticatorAdapter),
		fx.Provide(account.NewLoginRecorder),
		fx.Provide(
			fx.Annotate(transport.NewHandler, fx.As(new(platform.HTTPMounter)), fx.ResultTags(`group:"mounter"`)),
		),
	)
}

// AuthenticatorAdapter adapts account.Authenticator to satisfy auth.AccountAuthenticator.
type AuthenticatorAdapter struct {
	auth *account.Authenticator
}

// NewAuthenticatorAdapter creates an AuthenticatorAdapter.
func NewAuthenticatorAdapter(authenticator *account.Authenticator) *AuthenticatorAdapter {
	return &AuthenticatorAdapter{auth: authenticator}
}

// Authenticate delegates to the underlying account.Authenticator and converts
// the result to an auth.AuthenticatedAccount.
func (a *AuthenticatorAdapter) Authenticate(
	ctx context.Context, username, password string,
) (auth.AuthenticatedAccount, error) {
	acc, err := a.auth.Authenticate(ctx, username, password)
	if err != nil {
		return auth.AuthenticatedAccount{}, fmt.Errorf("authenticate: %w", err)
	}

	return auth.AuthenticatedAccount{
		ID:   int64(acc.ID()),
		Role: acc.Role().String(),
	}, nil
}
