package jwt_test

import (
	"errors"
	"strconv"
	"testing"
	"time"

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

func TestSigner_Sign(t *testing.T) {
	t.Parallel()

	validKey := "test-secret-key-that-is-long-enough"
	validity := 1 * time.Hour
	fakeNow := time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC)

	parser := jwt.NewParser(validKey, jwt.WithParserClock(func() time.Time {
		return fakeNow
	}))

	tests := []struct {
		name    string
		ids     idgen.Generator
		uid     int64
		role    string
		wantErr bool
	}{
		{
			name:    "success",
			ids:     &mockIDGen{id: 123456789, err: nil},
			uid:     1,
			role:    "admin",
			wantErr: false,
		},
		{
			name:    "id_generator_error",
			ids:     &mockIDGen{id: 0, err: errBroken},
			uid:     1,
			role:    "admin",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			signer := jwt.NewSigner(validKey, validity, tt.ids, jwt.WithSignerClock(func() time.Time {
				return fakeNow
			}))

			gotToken, err := signer.Sign(tt.uid, tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("Signer.Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if gotToken == "" {
				t.Fatal("Signer.Sign() returned empty token, expected non-empty")
			}

			claims, parseErr := parser.Parse(gotToken)
			if parseErr != nil {
				t.Fatalf("Failed to parse just issued token: %v", parseErr)
			}

			wantSub := strconv.FormatInt(tt.uid, 10)
			if claims.Subject != wantSub {
				t.Errorf("claims.Subject = %v, want %v", claims.Subject, wantSub)
			}

			if claims.Role != tt.role {
				t.Errorf("claims.Role = %v, want %v", claims.Role, tt.role)
			}

			if claims.ID != "123456789" {
				t.Errorf("claims.ID = %v, want %v", claims.ID, "123456789")
			}

			if !claims.IssuedAt.Equal(fakeNow) {
				t.Errorf("claims.IssuedAt = %v, want %v", claims.IssuedAt, fakeNow)
			}
		})
	}
}

//nolint:funlen // For testing only.
func TestParser_Parse(t *testing.T) {
	t.Parallel()

	validKey := "test-secret-key-that-is-long-enough"
	wrongKey := "wrong-key"
	validity := 1 * time.Hour
	fakeNow := time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC)

	signer := jwt.NewSigner(validKey, validity, &mockIDGen{id: 1, err: nil}, jwt.WithSignerClock(func() time.Time {
		return fakeNow
	}))
	validToken, err := signer.Sign(10, "user")
	if err != nil {
		t.Fatalf("Failed to sign valid token for test setup: %v", err)
	}

	expiredSigner := jwt.NewSigner(validKey, validity, &mockIDGen{id: 2, err: nil}, jwt.WithSignerClock(func() time.Time {
		return fakeNow.Add(-2 * time.Hour)
	}))
	expiredToken, err := expiredSigner.Sign(20, "expired_user")
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
		wantRole    string
	}{
		{
			name:        "success",
			tokenString: validToken,
			parser:      jwt.NewParser(validKey, jwt.WithParserClock(func() time.Time { return fakeNow })),
			wantErr:     nil,
			wantUID:     "10",
			wantRole:    "user",
		},
		{
			name:        "expired_token",
			tokenString: expiredToken,
			parser:   jwt.NewParser(validKey, jwt.WithParserClock(func() time.Time { return fakeNow })),
			wantErr:  jwt.ErrTokenExpired,
			wantUID:  "",
			wantRole: "",
		},
		{
			name:        "invalid_signature",
			tokenString: validToken,
			parser:      jwt.NewParser(wrongKey, jwt.WithParserClock(func() time.Time { return fakeNow })),
			wantErr:     jwt.ErrInvalidToken,
			wantUID:     "",
			wantRole:    "",
		},
		{
			name:        "malformed_token",
			tokenString: malformedToken,
			parser:      jwt.NewParser(validKey, jwt.WithParserClock(func() time.Time { return fakeNow })),
			wantErr:     jwt.ErrInvalidToken,
			wantUID:     "",
			wantRole:    "",
		},
		{
			name:        "empty_token",
			tokenString: "",
			parser:      jwt.NewParser(validKey, jwt.WithParserClock(func() time.Time { return fakeNow })),
			wantErr:     jwt.ErrInvalidToken,
			wantUID:     "",
			wantRole:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			claims, err := tt.parser.Parse(tt.tokenString)

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

			if claims.Role != tt.wantRole {
				t.Errorf("claims.Role = %v, want %v", claims.Role, tt.wantRole)
			}
		})
	}
}
