package jwt_test

import (
	"errors"
	"testing"
	"time"

	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
	"github.com/CubeLitBlade/community-v2/backend/internal/jwt"
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

var _ idgen.Generator = (*stubIDGen)(nil)

func TestIssuer_Issue(t *testing.T) {
	t.Parallel()

	cfg := &jwt.Config{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := jwt.NewIssuer(cfg, &stubIDGen{id: 42, err: nil})

	token, err := issuer.Issue(7, "member")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if token == "" {
		t.Fatal("Issue() returned empty token")
	}

	claims, err := jwt.Parse(token, []byte(cfg.Key))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	assertClaims(t, claims)
}

func TestIssuer_Issue_IDGenError(t *testing.T) {
	t.Parallel()

	cfg := &jwt.Config{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := jwt.NewIssuer(cfg, &stubIDGen{id: 0, err: errIDGen})

	_, err := issuer.Issue(1, "member")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewIssuer(t *testing.T) {
	t.Parallel()

	cfg := &jwt.Config{
		Key:      testJWTKey,
		Issuer:   testIssuer,
		Validity: 24 * time.Hour,
	}

	issuer := jwt.NewIssuer(cfg, &stubIDGen{id: 1, err: nil})
	if issuer == nil {
		t.Fatal("NewIssuer() returned nil")
	}
}

func TestParse_InvalidToken(t *testing.T) {
	t.Parallel()

	_, err := jwt.Parse("invalid-token", []byte(testJWTKey))
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func assertClaims(t *testing.T, claims *jwt.Claims) {
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
