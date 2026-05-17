package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AccountAuthenticator interface {
	Authenticate(
		ctx context.Context, username, password string,
	) (*account.Account, error)
}

type Login struct {
	Auth   AccountAuthenticator
	Issuer *JWTIssuer
}

func NewLogin(authenticator AccountAuthenticator, issuer *JWTIssuer) *Login {
	return &Login{
		Auth:   authenticator,
		Issuer: issuer,
	}
}

func (l *Login) Execute(
	ctx context.Context, username, password string,
) (string, error) {
	acc, err := l.Auth.Authenticate(ctx, username, password)
	if err != nil {
		if errors.Is(err, account.ErrInvalidCredentials) {
			return "", ErrInvalidCredentials
		}

		return "", err
	}

	jwt, err := l.Issuer.Issue(acc.ID(), acc.Role())
	if err != nil {
		return "", fmt.Errorf("issue token: %w", err)
	}

	return jwt, nil
}
