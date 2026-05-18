package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/CubeLitBlade/community-v2/backend/internal/jwt"
)

// ErrInvalidCredentials is returned when login fails due to invalid credentials.
var ErrInvalidCredentials = errors.New("invalid credentials")

// AccountAuthenticator authenticates a user by username and password.
type AccountAuthenticator interface {
	Authenticate(
		ctx context.Context, username, password string,
	) (*account.Account, error)
}

// Login executes the login flow: authenticate credentials and issue a JWT.
type Login struct {
	auth   AccountAuthenticator
	issuer *jwt.JWT
}

// Session holds the result of a successful login: a signed JWT and basic user info.
type Session struct {
	Token string
	ID    int64
	Role  string
}

// NewLogin creates a new Login service.
func NewLogin(authenticator AccountAuthenticator, issuer *jwt.JWT) *Login {
	return &Login{
		auth:   authenticator,
		issuer: issuer,
	}
}

// Execute authenticates the given credentials and returns a signed JWT.
func (l *Login) Execute(
	ctx context.Context, username, password string,
) (Session, error) {
	acc, err := l.auth.Authenticate(ctx, username, password)
	if err != nil {
		if errors.Is(err, account.ErrInvalidCredentials) {
			return Session{}, ErrInvalidCredentials
		}

		return Session{}, fmt.Errorf("authenticate: %w", err)
	}

	token, err := l.issuer.Issue(int64(acc.ID()), acc.Role().String())
	if err != nil {
		return Session{}, fmt.Errorf("issue token: %w", err)
	}

	return Session{
		Token: token,
		ID:    int64(acc.ID()),
		Role:  acc.Role().String(),
	}, nil
}
