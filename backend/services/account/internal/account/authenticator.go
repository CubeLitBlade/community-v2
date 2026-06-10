package account

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/shared"
	"gorm.io/gorm"
)

// AuthenticatorOption defines a functional option for configuring an Authenticator.
type AuthenticatorOption func(*Authenticator)

// ByUsernameFinder finds an account by username.
type ByUsernameFinder interface {
	FindByUsername(ctx context.Context, db *gorm.DB, username string) (*Account, error)
}

// Authenticator verifies username/password credentials against stored accounts.
type Authenticator struct {
	finder ByUsernameFinder
	db     *gorm.DB
	clock  func() time.Time
	logger *slog.Logger
}

// NewAuthenticator creates a new Authenticator.
func NewAuthenticator(finder ByUsernameFinder, db *gorm.DB,
	logger *slog.Logger, opts ...AuthenticatorOption,
) *Authenticator {
	logger = logger.With(
		slog.String("service", "account"),
		slog.String("component", "account/authenticator"),
	)

	authenticator := &Authenticator{
		finder: finder,
		db:     db,
		clock:  time.Now,
		logger: logger,
	}

	for _, opt := range opts {
		opt(authenticator)
	}

	return authenticator
}

// Authenticate verifies the given username and password and returns the matching account.
func (a *Authenticator) Authenticate(
	ctx context.Context, username, password string,
) (*shared.AuthenticatedAccount, error) {
	acc, err := a.finder.FindByUsername(ctx, a.db, username)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			a.logger.WarnContext(ctx, "authentication failed: account not found.",
				slog.String("username", username),
			)

			return nil, ErrInvalidCredentials // hide details
		}

		return nil, fmt.Errorf("find by username: %w", err)
	}

	if acc == nil {
		return nil, ErrInvalidCredentials
	}

	if !acc.VerifyPassword(password) {
		a.logger.WarnContext(ctx, "authentication failed: username or password is incorrect",
			slog.String("username", username),
		)

		return nil, ErrInvalidCredentials
	}

	return &shared.AuthenticatedAccount{
		ID:   acc.ID(),
		Role: acc.Role().String(),
	}, nil
}

// WithAuthenticatorClock sets the clock function for the Authenticator.
func WithAuthenticatorClock(clock func() time.Time) AuthenticatorOption {
	return func(a *Authenticator) {
		a.clock = clock
	}
}
