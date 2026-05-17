package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/CubeLitBlade/community-v2/backend/internal/auth"
	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
)

const (
	testJWTKey  = "super-secret-key-at-least-32-bytes!"
	testIssuer  = "community-v2"
	testSubject = "7"
	testTokenID = "42"
)

var errIDGen = errors.New("id generation failed")

type stubIDGen struct {
	id  int64
	err error
}

func (g *stubIDGen) NextID() (int64, error) {
	return g.id, g.err
}

// compile-time check
var _ idgen.Generator = (*stubIDGen)(nil)

func TestJWTIssuer_Issue(t *testing.T) {
	t.Parallel()

	cfg := &auth.JWTConfig{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := auth.NewJWTIssuer(cfg, &stubIDGen{id: 42, err: nil})

	token, err := issuer.Issue(account.ID(7), account.RoleMember)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if token == "" {
		t.Fatal("Issue() returned empty token")
	}

	claims := parseToken(t, token, cfg.Key)
	assertClaims(t, claims)
}

func TestJWTIssuer_Issue_IDGenError(t *testing.T) {
	t.Parallel()

	cfg := &auth.JWTConfig{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := auth.NewJWTIssuer(cfg, &stubIDGen{id: 0, err: errIDGen})

	_, err := issuer.Issue(account.ID(1), account.RoleMember)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewJWTIssuer(t *testing.T) {
	t.Parallel()

	cfg := &auth.JWTConfig{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := auth.NewJWTIssuer(cfg, &stubIDGen{id: 1, err: nil})
	if issuer == nil {
		t.Fatal("NewJWTIssuer() returned nil")
	}
}

func parseToken(t *testing.T, token, key string) *auth.Claims {
	t.Helper()

	claims := &auth.Claims{
		Role:             "",
		RegisteredClaims: jwt.RegisteredClaims{},
	}

	parsed, err := jwt.ParseWithClaims(
		token, claims,
		func(_ *jwt.Token) (any, error) {
			return []byte(key), nil
		},
	)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if !parsed.Valid {
		t.Fatal("parsed token is not valid")
	}

	return claims
}

func assertClaims(t *testing.T, claims *auth.Claims) {
	t.Helper()

	now := time.Now()

	if claims.Role != "member" {
		t.Errorf("Role = %q, want %q", claims.Role, "member")
	}

	if claims.Subject != testSubject {
		t.Errorf("Subject = %q, want %q", claims.Subject, testSubject)
	}

	if claims.Issuer != testIssuer {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, testIssuer)
	}

	if claims.ID != testTokenID {
		t.Errorf("ID = %q, want %q", claims.ID, testTokenID)
	}

	if claims.ExpiresAt == nil {
		t.Fatal("ExpiresAt is nil")
	}

	if claims.ExpiresAt.Before(now) {
		t.Error("token has already expired")
	}

	if claims.IssuedAt == nil {
		t.Fatal("IssuedAt is nil")
	}
}
