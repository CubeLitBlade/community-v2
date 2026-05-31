package account

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

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
	if logger == nil {
		logger = slog.Default()
		logger.Warn("No logger provided while creating authenticator, using default logger.")
	}

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
func (a *Authenticator) Authenticate(ctx context.Context, username, password string) (*Account, error) {
	acc, err := a.finder.FindByUsername(ctx, a.db, username)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			a.logger.Warn("Authentication failed: account not found.", "username", username)

			return nil, ErrInvalidCredentials // hide details
		}

		return nil, fmt.Errorf("find by username: %w", err)
	}

	if acc == nil {
		return nil, ErrInvalidCredentials
	}

	if !acc.VerifyPassword(password) {
		a.logger.Warn("Authentication failed: username or password is incorrect.", "username", username)

		return nil, ErrInvalidCredentials
	}

	return acc, nil
}

// WithAuthenticatorClock sets the clock function for the Authenticator.
func WithAuthenticatorClock(clock func() time.Time) AuthenticatorOption {
	return func(a *Authenticator) {
		a.clock = clock
	}
}
