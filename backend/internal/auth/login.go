package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
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
	issuer *JWTIssuer
}

// NewLogin creates a new Login service.
func NewLogin(authenticator AccountAuthenticator, issuer *JWTIssuer) *Login {
	return &Login{
		auth:   authenticator,
		issuer: issuer,
	}
}

// Execute authenticates the given credentials and returns a signed JWT.
func (l *Login) Execute(
	ctx context.Context, username, password string,
) (string, error) {
	acc, err := l.auth.Authenticate(ctx, username, password)
	if err != nil {
		if errors.Is(err, account.ErrInvalidCredentials) {
			return "", ErrInvalidCredentials
		}

		return "", fmt.Errorf("authenticate: %w", err)
	}

	jwt, err := l.issuer.Issue(acc.ID(), acc.Role())
	if err != nil {
		return "", fmt.Errorf("issue token: %w", err)
	}

	return jwt, nil
}
