package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/netip"
	"strconv"

	"github.com/cubelitblade/community-v2/backend/pkg/common/jwt"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/contracts"
)

// ErrInvalidCredentials is returned when login fails due to invalid credentials.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Authenticator authenticates a user by username and password.
type Authenticator interface {
	Authenticate(ctx context.Context, username, password string) (*contracts.AuthenticatedAccount, error)
}

// LoginRecorder records a successful login event.
type LoginRecorder interface {
	Record(ctx context.Context, accountID int64, ip netip.Addr) error
}

// Login executes the login flow: authenticate credentials and issue a JWT.
type Login struct {
	auth     Authenticator
	recorder LoginRecorder
	signer   *jwt.Signer
	ttl      contracts.TTLProvider
}

// Credentials holds the result of a successful login: a signed JWT and basic user info.
type Credentials struct {
	Token string
	ID    int64
	Role  string
}

// NewLogin creates a new Login service.
func NewLogin(auth Authenticator, signer *jwt.Signer, recorder LoginRecorder, ttl contracts.TTLProvider) *Login {
	return &Login{
		auth:     auth,
		signer:   signer,
		recorder: recorder,
		ttl:      ttl,
	}
}

// Execute authenticates the given credentials and returns a signed JWT.
func (l *Login) Execute(ctx context.Context, username, password string, ipaddr netip.Addr) (Credentials, error) {
	acc, err := l.auth.Authenticate(ctx, username, password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return Credentials{}, ErrInvalidCredentials
		}

		return Credentials{}, fmt.Errorf("authenticate: %w", err)
	}

	regClaims, err := l.signer.NewClaims(strconv.FormatInt(acc.ID, 10), l.ttl.AccessTokenTTL())
	if err != nil {
		return Credentials{}, fmt.Errorf("sign new claims: %w", err)
	}

	token, err := l.signer.Sign(contracts.AccessTokenClaims{
		RegisteredClaims: regClaims,
		Role:             acc.Role,
	})
	if err != nil {
		return Credentials{}, fmt.Errorf("sign token: %w", err)
	}

	if err := l.recorder.Record(ctx, acc.ID, ipaddr); err != nil {
		log.Printf("failed to record access record: %v", err)
	}

	return Credentials{
		Token: token,
		ID:    acc.ID,
		Role:  acc.Role,
	}, nil
}
