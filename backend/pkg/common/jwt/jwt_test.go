package jwt_test

import (
	"errors"
	"strconv"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/pkg/common/jwt"
)

var errBroken = errors.New("snowflake broken")

type mockIDGen struct {
	id  int64
	err error
}

func (m *mockIDGen) NextID() (int64, error) {
	return m.id, m.err
}

//nolint:funlen // For testing only.
func TestSigner_Sign(t *testing.T) {
	t.Parallel()

	validKey := []byte("test-secret-key-that-is-long-enough")
	defaultTTL := 1 * time.Hour
	fakeNow := time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC)

	parser := jwt.NewParser(validKey, jwt.WithParserClock(func() time.Time {
		return fakeNow
	}))

	tests := []struct {
		name    string
		ids     idgen.Generator
		uid     int64
		ttl     time.Duration
		wantErr bool
	}{
		{
			name:    "success",
			ids:     &mockIDGen{id: 123456789, err: nil},
			uid:     1,
			ttl:     defaultTTL,
			wantErr: false,
		},
		{
			name:    "id_generator_error",
			ids:     &mockIDGen{id: 0, err: errBroken},
			uid:     1,
			ttl:     defaultTTL,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			signer := jwt.NewSigner(validKey, tt.ids, jwt.WithSignerClock(func() time.Time {
				return fakeNow
			}))

			claims, err := signer.NewClaims(strconv.FormatInt(tt.uid, 10), tt.ttl)
			if (err != nil) != tt.wantErr {
				t.Errorf("Signer.NewClaims() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			gotToken, err := signer.Sign(claims)
			if err != nil {
				t.Errorf("Signer.Sign() error = %v", err)
				return
			}

			if gotToken == "" {
				t.Fatal("Signer.Sign() returned empty token, expected non-empty")
			}

			var parsedClaims jwtlib.RegisteredClaims

			parseErr := parser.Parse(gotToken, &parsedClaims)
			if parseErr != nil {
				t.Fatalf("Failed to parse just issued token: %v", parseErr)
			}

			wantSub := strconv.FormatInt(tt.uid, 10)
			if parsedClaims.Subject != wantSub {
				t.Errorf("parsedClaims.Subject = %v, want %v", parsedClaims.Subject, wantSub)
			}

			if parsedClaims.ID != "123456789" {
				t.Errorf("parsedClaims.ID = %v, want %v", parsedClaims.ID, "123456789")
			}

			if !parsedClaims.IssuedAt.Equal(fakeNow) {
				t.Errorf("parsedClaims.IssuedAt = %v, want %v", parsedClaims.IssuedAt, fakeNow)
			}
		})
	}
}

//nolint:funlen // For testing only.
func TestParser_Parse(t *testing.T) {
	t.Parallel()

	validKey := []byte("test-secret-key-that-is-long-enough")
	wrongKey := []byte("wrong-key")
	defaultTTL := 1 * time.Hour
	fakeNow := time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC)

	signer := jwt.NewSigner(validKey, &mockIDGen{id: 1, err: nil}, jwt.WithSignerClock(func() time.Time {
		return fakeNow
	}))

	validClaims, err := signer.NewClaims("10", defaultTTL)
	if err != nil {
		t.Fatalf("Failed to create claims for test setup: %v", err)
	}

	validToken, err := signer.Sign(validClaims)
	if err != nil {
		t.Fatalf("Failed to sign valid token for test setup: %v", err)
	}

	expiredSigner := jwt.NewSigner(validKey, &mockIDGen{id: 2, err: nil}, jwt.WithSignerClock(func() time.Time {
		return fakeNow.Add(-2 * time.Hour)
	}))

	expiredClaims, err := expiredSigner.NewClaims("20", defaultTTL)
	if err != nil {
		t.Fatalf("Failed to create expired claims for test setup: %v", err)
	}

	expiredToken, err := expiredSigner.Sign(expiredClaims)
	if err != nil {
		t.Fatalf("Failed to sign expired token for test setup: %v", err)
	}

	//nolint:gosec // Fake token for test.
	malformedToken := "this.is.not.a.jwt"

	tests := []struct {
		name        string
		tokenString string
		parser      *jwt.Parser
		wantErr     error
		wantUID     string
	}{
		{
			name:        "success",
			tokenString: validToken,
			parser:      jwt.NewParser(validKey, jwt.WithParserClock(func() time.Time { return fakeNow })),
			wantErr:     nil,
			wantUID:     "10",
		},
		{
			name:        "expired_token",
			tokenString: expiredToken,
			parser:      jwt.NewParser(validKey, jwt.WithParserClock(func() time.Time { return fakeNow })),
			wantErr:     jwt.ErrTokenExpired,
			wantUID:     "",
		},
		{
			name:        "invalid_signature",
			tokenString: validToken,
			parser:      jwt.NewParser(wrongKey, jwt.WithParserClock(func() time.Time { return fakeNow })),
			wantErr:     jwt.ErrInvalidToken,
			wantUID:     "",
		},
		{
			name:        "malformed_token",
			tokenString: malformedToken,
			parser:      jwt.NewParser(validKey, jwt.WithParserClock(func() time.Time { return fakeNow })),
			wantErr:     jwt.ErrInvalidToken,
			wantUID:     "",
		},
		{
			name:        "empty_token",
			tokenString: "",
			parser:      jwt.NewParser(validKey, jwt.WithParserClock(func() time.Time { return fakeNow })),
			wantErr:     jwt.ErrInvalidToken,
			wantUID:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var claims jwtlib.RegisteredClaims

			err := tt.parser.Parse(tt.tokenString, &claims)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Parser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				return
			}

			if claims.Subject != tt.wantUID {
				t.Errorf("claims.Subject = %v, want %v", claims.Subject, tt.wantUID)
			}
		})
	}
}
