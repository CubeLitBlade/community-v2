package account

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// ErrInvalidCredentials is returned when authentication fails due to mismatched credentials.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ByUsernameFinder looks up an account by its username.
type ByUsernameFinder interface {
	FindByUsername(ctx context.Context, username string) (*Account, error)
}

// Authenticator verifies username/password credentials against stored accounts.
type Authenticator struct {
	finder ByUsernameFinder
	now    func() time.Time
	logger *slog.Logger
}

// NewAuthenticator creates a new Authenticator.
func NewAuthenticator(
	finder ByUsernameFinder, logger *slog.Logger,
) *Authenticator {
	return &Authenticator{
		finder: finder,
		now:    time.Now,
		logger: logger,
	}
}

// Authenticate verifies the given username and password and returns the matching account.
func (a *Authenticator) Authenticate(
	ctx context.Context, username, password string,
) (*Account, error) {
	acc, err := a.finder.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			return nil, ErrInvalidCredentials // hide details
		}
		return nil, fmt.Errorf("find by username: %w", err)
	}

	if acc == nil {
		return nil, ErrInvalidCredentials
	}

	if !acc.VerifyPassword(password) {
		return nil, ErrInvalidCredentials
	}

	return acc, nil
}
