package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/netip"

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

// LoginRecorder records a successful login event.
type LoginRecorder interface {
	Record(ctx context.Context, acc *account.Account, ip netip.Addr) error
}

// Login executes the login flow: authenticate credentials and issue a JWT.
type Login struct {
	auth     AccountAuthenticator
	recorder LoginRecorder
	issuer   *jwt.JWT
}

// Credentials holds the result of a successful login: a signed JWT and basic user info.
type Credentials struct {
	Token string
	ID    int64
	Role  string
}

// NewLogin creates a new Login service.
func NewLogin(
	auth AccountAuthenticator, issuer *jwt.JWT, recorder LoginRecorder,
) *Login {
	return &Login{
		auth:     auth,
		issuer:   issuer,
		recorder: recorder,
	}
}

// Execute authenticates the given credentials and returns a signed JWT.
func (l *Login) Execute(
	ctx context.Context, username, password string, ipaddr netip.Addr,
) (Credentials, error) {
	acc, err := l.auth.Authenticate(ctx, username, password)
	if err != nil {
		if errors.Is(err, account.ErrInvalidCredentials) {
			return Credentials{}, ErrInvalidCredentials
		}

		return Credentials{}, fmt.Errorf("authenticate: %w", err)
	}

	token, err := l.issuer.Issue(int64(acc.ID()), acc.Role().String())
	if err != nil {
		return Credentials{}, fmt.Errorf("issue token: %w", err)
	}

	if err := l.recorder.Record(ctx, acc, ipaddr); err != nil {
		log.Printf("failed to record access record: %v", err)
	}

	return Credentials{
		Token: token,
		ID:    int64(acc.ID()),
		Role:  acc.Role().String(),
	}, nil
}
