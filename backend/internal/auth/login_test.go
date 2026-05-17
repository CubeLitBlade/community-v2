package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/CubeLitBlade/community-v2/backend/internal/auth"
)

var errAuthTimeout = errors.New("db timeout")

type stubAuthenticator struct {
	acc *account.Account
	err error
}

func (a *stubAuthenticator) Authenticate(
	_ context.Context, _, _ string,
) (*account.Account, error) {
	if a.err != nil {
		return nil, a.err
	}

	return a.acc, nil
}

// compile-time check
var _ auth.AccountAuthenticator = (*stubAuthenticator)(nil)

func makeTestAccount(t *testing.T) *account.Account {
	t.Helper()

	acc, err := account.Register(
		1, "testuser", "this-is-a-valid-password",
		time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	return &acc
}

func TestLogin_Execute_Success(t *testing.T) {
	t.Parallel()

	cfg := &auth.JWTConfig{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := auth.NewJWTIssuer(cfg, &stubIDGen{id: 99, err: nil})
	acc := makeTestAccount(t)
	login := auth.NewLogin(
		&stubAuthenticator{acc: acc, err: nil},
		issuer,
	)

	token, err := login.Execute(
		context.Background(), "testuser", "this-is-a-valid-password",
	)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if token == "" {
		t.Fatal("Execute() returned empty token")
	}
}

func TestLogin_Execute_InvalidCredentials(t *testing.T) {
	t.Parallel()

	cfg := &auth.JWTConfig{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := auth.NewJWTIssuer(cfg, &stubIDGen{id: 99, err: nil})
	login := auth.NewLogin(
		&stubAuthenticator{acc: nil, err: account.ErrInvalidCredentials},
		issuer,
	)

	_, err := login.Execute(
		context.Background(), "bad", "password",
	)
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("error = %v, want %v",
			err, auth.ErrInvalidCredentials)
	}
}

func TestLogin_Execute_UnexpectedError(t *testing.T) {
	t.Parallel()

	cfg := &auth.JWTConfig{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := auth.NewJWTIssuer(cfg, &stubIDGen{id: 99, err: nil})
	login := auth.NewLogin(
		&stubAuthenticator{acc: nil, err: errAuthTimeout},
		issuer,
	)

	_, err := login.Execute(
		context.Background(), "any", "password",
	)
	if !errors.Is(err, errAuthTimeout) {
		t.Errorf("error = %v, want %v", err, errAuthTimeout)
	}
}

func TestNewLogin(t *testing.T) {
	t.Parallel()

	cfg := &auth.JWTConfig{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := auth.NewJWTIssuer(cfg, &stubIDGen{id: 1, err: nil})
	login := auth.NewLogin(
		&stubAuthenticator{acc: nil, err: nil},
		issuer,
	)
	if login == nil {
		t.Fatal("NewLogin() returned nil")
	}
}
