package account

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type ByUsernameFinder interface {
	FindByUsername(ctx context.Context, username string) (*Account, error)
}

type Authenticator struct {
	finder ByUsernameFinder
	now    func() time.Time
	logger *slog.Logger
}

func NewAuthenticator(
	finder ByUsernameFinder, logger *slog.Logger,
) *Authenticator {
	return &Authenticator{
		finder: finder,
		now:    time.Now,
		logger: logger,
	}
}

func (a *Authenticator) Authenticate(
	ctx context.Context, username, password string,
) (*Account, error) {
	acc, err := a.finder.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			return nil, ErrInvalidCredentials // hide details
		}
		return nil, err
	}

	if acc == nil {
		return nil, ErrInvalidCredentials
	}

	if !acc.VerifyPassword(password) {
		return nil, ErrInvalidCredentials
	}

	return acc, nil
}
