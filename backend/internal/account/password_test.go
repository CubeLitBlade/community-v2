package account_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"golang.org/x/crypto/bcrypt"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr error
	}{
		{
			name:    "rejects empty password",
			raw:     "",
			wantErr: account.ErrPasswordEmpty,
		},
		{
			name:    "rejects password shorter than minimum length",
			raw:     "a",
			wantErr: account.ErrPasswordTooShort,
		},
		{
			name: "accepts valid password",
			raw:  "this-is-a-valid-password",
		},
		{
			name:    "rejects password longer than bcrypt byte limit",
			raw:     strings.Repeat("a", account.MaxPasswordBytes+1),
			wantErr: account.ErrPasswordTooLong,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := account.ValidatePassword(tt.raw)
			if gotErr != nil {
				if tt.wantErr == nil {
					t.Errorf("ValidatePassword() failed: %v", gotErr)
				}

				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("ValidatePassword() error = %v, want %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr != nil {
				t.Fatal("ValidatePassword() succeeded unexpectedly")
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	raw := "this-is-a-valid-password"

	hash, err := account.HashPassword(raw)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if hash.Value() == "" {
		t.Fatal("HashPassword() returned empty hash")
	}

	if hash.Value() == raw {
		t.Fatal("HashPassword() returned raw password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash.Value()), []byte(raw))
	if err != nil {
		t.Fatalf("generated hash does not match raw password: %v", err)
	}
}

func TestPasswordMatches(t *testing.T) {
	raw := "this-is-a-valid-password"

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword() failed: %v", err)
	}

	hash, err := account.PasswordHashFromStorage(string(hashedBytes))
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
			raw:  "",
			want: false,
		},
		{
			name: "does not match password longer than bcrypt byte limit",
			raw:  strings.Repeat("a", account.MaxPasswordBytes+1),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := account.PasswordMatches(hash, tt.raw)

			if got != tt.want {
				t.Errorf("PasswordMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordHashAndMatchWorkTogether(t *testing.T) {
	raw := "this-is-a-valid-password"

	hash, err := account.HashPassword(raw)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if !account.PasswordMatches(hash, raw) {
		t.Fatal("PasswordMatches() returned false for password hashed by HashPassword()")
	}
}
