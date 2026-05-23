// Package setup provides constructors that wire together account domain services,
// persistence, and HTTP transport.
package setup

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
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
		fx.Provide(postgres.NewReader),
		fx.Provide(postgres.NewWriter),
		fx.Provide(NewAuthenticator),
		fx.Provide(NewAuthenticatorAdapter),
		fx.Provide(NewLoginRecorder),
		fx.Provide(
			fx.Annotate(NewHandler, fx.As(new(platform.HTTPMounter)), fx.ResultTags(`group:"mounter"`)),
		),
	)
}

// NewHandler creates the account HTTP handler with all dependencies wired.
func NewHandler(
	reader *postgres.Reader, writer *postgres.Writer, ids idgen.Generator, logger *slog.Logger,
) *transport.Handler {
	registrar := account.NewRegistrar(ids, writer, logger)
	finder := account.NewProfileFinder(reader, logger)

	return transport.NewHandler(transport.Deps{
		Registrar:     registrar,
		ProfileFinder: finder,
		Logger:        logger,
	})
}

// NewAuthenticator creates the account authenticator for cross-module injection.
func NewAuthenticator(reader *postgres.Reader, logger *slog.Logger) *account.Authenticator {
	return account.NewAuthenticator(reader, logger)
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

// NewLoginRecorder creates an account.LoginRecorder that uses the given reader
// to look up accounts, the writer to persist login events,
// and time.Now as the clock source.
func NewLoginRecorder(
	reader *postgres.Reader, writer *postgres.Writer, logger *slog.Logger,
) *account.LoginRecorder {
	return account.NewLoginRecorder(time.Now, reader, writer, logger)
}
