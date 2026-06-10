package auth_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/pkg/common/jwt"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/auth"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/shared"
)

const (
	testJWTKey   = "super-secret-key-at-least-32-bytes!"
	testIssuer   = "community-v2"
	testValidity = 24 * time.Hour
)

var (
	errAuthTimeout = errors.New("db timeout")
	errBrokenIDGen = errors.New("snowflake broken")
)

type stubIDGen struct {
	id  int64
	err error
}

func (g *stubIDGen) NextID() (int64, error) {
	return g.id, g.err
}

// Compile-time check.
var _ idgen.Generator = (*stubIDGen)(nil)

type stubAuthenticator struct {
	acc *shared.AuthenticatedAccount
	err error
}

func (a *stubAuthenticator) Authenticate(
	_ context.Context, _, _ string,
) (*shared.AuthenticatedAccount, error) {
	if a.err != nil {
		return nil, a.err
	}

	return a.acc, nil
}

// Compile-time check.
var _ auth.Authenticator = (*stubAuthenticator)(nil)

type stubLoginRecorder struct{}

func (r *stubLoginRecorder) Record(
	_ context.Context, _ int64, _ netip.Addr,
) error {
	return nil
}

// Compile-time check.
var _ auth.LoginRecorder = (*stubLoginRecorder)(nil)

func TestLogin_Execute(t *testing.T) {
	t.Parallel()

	ipaddr := netip.MustParseAddr("127.0.0.1")

	tests := []struct {
		name      string
		authAcc   shared.AuthenticatedAccount
		authErr   error
		idGenErr  error
		wantErr   error
		wantToken bool
	}{
		{
			name:      "success",
			authAcc:   shared.AuthenticatedAccount{ID: 1, Role: "member"},
			authErr:   nil,
			idGenErr:  nil,
			wantErr:   nil,
			wantToken: true,
		},
		{
			name:      "invalid_credentials",
			authAcc:   shared.AuthenticatedAccount{ID: 0, Role: ""},
			authErr:   auth.ErrInvalidCredentials,
			idGenErr:  nil,
			wantErr:   auth.ErrInvalidCredentials,
			wantToken: false,
		},
		{
			name:      "unexpected_auth_error",
			authAcc:   shared.AuthenticatedAccount{ID: 0, Role: ""},
			authErr:   errAuthTimeout,
			idGenErr:  nil,
			wantErr:   errAuthTimeout,
			wantToken: false,
		},
		{
			name:      "id_generator_error",
			authAcc:   shared.AuthenticatedAccount{ID: 1, Role: "member"},
			authErr:   nil,
			idGenErr:  errBrokenIDGen,
			wantErr:   errBrokenIDGen,
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			signer := jwt.NewSigner([]byte(testJWTKey), &stubIDGen{id: 99, err: tt.idGenErr}, jwt.WithSignerIssuer(testIssuer))

			login := auth.NewLogin(
				&stubAuthenticator{acc: &tt.authAcc, err: tt.authErr},
				signer,
				&stubLoginRecorder{},
				testValidity,
			)

			session, err := login.Execute(
				context.Background(),
				"test_user",
				"password",
				ipaddr,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantToken {
				if session.Token == "" {
					t.Error("Execute() returned empty token, expected non-empty")
				}
			}
		})
	}
}
