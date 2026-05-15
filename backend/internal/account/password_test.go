package account_test

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
)

const (
	validPassword = "this-is-a-valid-password"
	empty         = ""
	padChar       = "a"
)

//nolint:gochecknoglobals // test data requiring runtime initialization
var tooLongPassword = strings.Repeat(padChar, account.MaxPasswordBytes+1)

func assertError(t *testing.T, gotErr, wantErr error) {
	t.Helper()

	if wantErr == nil {
		if gotErr != nil {
			t.Errorf("unexpected error: %v", gotErr)
		}

		return
	}

	if gotErr == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(gotErr, wantErr) {
		t.Errorf("error = %v, want %v", gotErr, wantErr)
	}
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		name    string
		raw     string
	}{
		{
			name:    "rejects empty password",
			raw:     empty,
			wantErr: account.ErrPasswordEmpty,
		},
		{
			name:    "rejects password shorter than minimum length",
			raw:     padChar,
			wantErr: account.ErrPasswordTooShort,
		},
		{
			name:    "accepts valid password",
			raw:     validPassword,
			wantErr: nil,
		},
		{
			name:    "rejects password longer than bcrypt byte limit",
			raw:     tooLongPassword,
			wantErr: account.ErrPasswordTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotErr := account.ValidatePassword(tt.raw)
			assertError(t, gotErr, tt.wantErr)
		})
	}
}

func TestHashPassword(t *testing.T) {
	t.Parallel()

	raw := validPassword

	hash, err := account.HashPassword(raw)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if hash.Value() == empty {
		t.Fatal("HashPassword() returned empty hash")
	}

	if hash.Value() == raw {
		t.Fatal("HashPassword() returned raw password")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(hash.Value()), []byte(raw),
	)
	if err != nil {
		t.Fatalf("generated hash does not match raw password: %v", err)
	}
}

func TestPasswordMatches(t *testing.T) {
	t.Parallel()

	raw := validPassword

	hashedBytes, err := bcrypt.GenerateFromPassword(
		[]byte(raw), bcrypt.DefaultCost,
	)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword() failed: %v", err)
	}

	hash, err := account.PasswordHashFromStorage(
		string(hashedBytes),
	)
	if err != nil {
		t.Fatalf("PasswordHashFromStorage() failed: %v", err)
	}

	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{
			name: "matches correct password",
			raw:  raw,
			want: true,
		},
		{
			name: "does not match wrong password",
			raw:  "this-is-a-wrong-password",
			want: false,
		},
		{
			name: "does not match empty password",
			raw:  empty,
			want: false,
		},
		{
			name: "does not match password longer than bcrypt byte limit",
			raw:  tooLongPassword,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := account.PasswordMatches(hash, tt.raw)

			if got != tt.want {
				t.Errorf(
					"PasswordMatches() = %v, want %v",
					got, tt.want,
				)
			}
		})
	}
}

func TestPasswordHashAndMatchWorkTogether(t *testing.T) {
	t.Parallel()

	raw := validPassword

	hash, err := account.HashPassword(raw)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if !account.PasswordMatches(hash, raw) {
		t.Fatal(
			"PasswordMatches() returned false for hash from HashPassword()",
		)
	}
}
