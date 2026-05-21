package auth_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/CubeLitBlade/community-v2/backend/pkg/common/idgen"
	"github.com/CubeLitBlade/community-v2/backend/pkg/common/jwt"
	"github.com/CubeLitBlade/community-v2/backend/services/user/internal/auth"
)

const (
	testJWTKey = "super-secret-key-at-least-32-bytes!"
	testIssuer = "community-v2"
)

var errAuthTimeout = errors.New("db timeout")

type stubIDGen struct {
	id  int64
	err error
}

func (g *stubIDGen) NextID() (int64, error) {
	return g.id, g.err
}

var _ idgen.Generator = (*stubIDGen)(nil)

type stubAuthenticator struct {
	acc auth.AuthenticatedAccount
	err error
}

func (a *stubAuthenticator) Authenticate(
	_ context.Context, _, _ string,
) (auth.AuthenticatedAccount, error) {
	if a.err != nil {
		return auth.AuthenticatedAccount{ID: 0, Role: ""}, a.err
	}

	return a.acc, nil
}

type stubLoginRecorder struct{}

func (r *stubLoginRecorder) Record(
	_ context.Context, _ int64, _ netip.Addr,
) error {
	return nil
}

// compile-time checks
var (
	_ auth.AccountAuthenticator = (*stubAuthenticator)(nil)
	_ auth.LoginRecorder        = (*stubLoginRecorder)(nil)
)

func TestLogin_Execute_Success(t *testing.T) {
	t.Parallel()

	cfg := &jwt.Config{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := jwt.New(cfg, &stubIDGen{id: 99, err: nil})
	login := auth.NewLogin(
		&stubAuthenticator{
			acc: auth.AuthenticatedAccount{ID: 1, Role: "member"},
			err: nil,
		},
		issuer,
		&stubLoginRecorder{},
	)
	ipaddr := netip.MustParseAddr("127.0.0.1")

	session, err := login.Execute(
		context.Background(),
		"testuser",
		"this-is-a-valid-password",
		ipaddr,
	)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if session.Token == "" {
		t.Fatal("Execute() returned empty token")
	}
}

func TestLogin_Execute_InvalidCredentials(t *testing.T) {
	t.Parallel()

	cfg := &jwt.Config{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := jwt.New(cfg, &stubIDGen{id: 99, err: nil})
	login := auth.NewLogin(
		&stubAuthenticator{
			acc: auth.AuthenticatedAccount{ID: 0, Role: ""},
			err: auth.ErrInvalidCredentials,
		},
		issuer,
		&stubLoginRecorder{},
	)
	ip := netip.MustParseAddr("127.0.0.1")

	_, err := login.Execute(
		context.Background(), "bad", "password", ip,
	)
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("error = %v, want %v",
			err, auth.ErrInvalidCredentials)
	}
}

func TestLogin_Execute_UnexpectedError(t *testing.T) {
	t.Parallel()

	cfg := &jwt.Config{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := jwt.New(cfg, &stubIDGen{id: 99, err: nil})
	login := auth.NewLogin(
		&stubAuthenticator{
			acc: auth.AuthenticatedAccount{ID: 0, Role: ""},
			err: errAuthTimeout,
		},
		issuer,
		&stubLoginRecorder{},
	)
	ip := netip.MustParseAddr("127.0.0.1")

	_, err := login.Execute(
		context.Background(), "any", "password", ip,
	)
	if !errors.Is(err, errAuthTimeout) {
		t.Errorf("error = %v, want %v", err, errAuthTimeout)
	}
}

func TestNewLogin(t *testing.T) {
	t.Parallel()

	cfg := &jwt.Config{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := jwt.New(cfg, &stubIDGen{id: 1, err: nil})
	login := auth.NewLogin(
		&stubAuthenticator{
			acc: auth.AuthenticatedAccount{ID: 0, Role: ""},
			err: nil,
		},
		issuer,
		&stubLoginRecorder{},
	)
	if login == nil {
		t.Fatal("NewLogin() returned nil")
	}
}
