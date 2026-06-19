package account_test

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/account"
)

const (
	validPwd = "this-is-a-valid-password"
	emptyPwd = ""
	padChar  = "a"
)

//nolint:gochecknoglobals // test data requiring runtime initialization
var tooLongPassword = strings.Repeat(padChar, account.MaxPasswordBytes+1)

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		name    string
		raw     string
	}{
		{
			name:    "rejects empty password",
			raw:     emptyPwd,
			wantErr: account.ErrPasswordEmpty,
		},
		{
			name:    "rejects password shorter than minimum",
			raw:     padChar,
			wantErr: account.ErrPasswordTooShort,
		},
		{
			name:    "accepts valid password",
			raw:     validPwd,
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
			assertPasswordError(t, gotErr, tt.wantErr)
		})
	}
}

func TestHashPassword(t *testing.T) {
	t.Parallel()

	raw := validPwd

	hash, err := account.HashPassword(raw)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if hash.Value() == emptyPwd {
		t.Fatal("HashPassword() returned empty hash")
	}

	if hash.Value() == raw {
		t.Fatal("HashPassword() returned raw password")
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(hash.Value()), []byte(raw),
	); err != nil {
		t.Fatalf("generated hash does not match raw password: %v", err)
	}
}

func TestPasswordHashFromStorage(t *testing.T) {
	t.Parallel()

	t.Run("accepts valid hash string", func(t *testing.T) {
		t.Parallel()

		hash, err := account.PasswordHashFromStorage("some-hash")
		if err != nil {
			t.Fatalf("PasswordHashFromStorage() failed: %v", err)
		}

		if hash.Value() != "some-hash" {
			t.Errorf("Value() = %q, want %q", hash.Value(), "some-hash")
		}
	})

	t.Run("rejects empty hash string", func(t *testing.T) {
		t.Parallel()

		_, err := account.PasswordHashFromStorage("")
		if !errors.Is(err, account.ErrPasswordHashEmpty) {
			t.Errorf("error = %v, want %v", err, account.ErrPasswordHashEmpty)
		}
	})
}

func TestPasswordHashValue(t *testing.T) {
	t.Parallel()

	hash, err := account.HashPassword(validPwd)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if hash.Value() == emptyPwd {
		t.Fatal("Value() returned empty string")
	}
}

func TestPasswordHashString(t *testing.T) {
	t.Parallel()

	hash, err := account.HashPassword(validPwd)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if hash.String() != "[redacted]" {
		t.Errorf("String() = %q, want %q", hash.String(), "[redacted]")
	}
}

func assertPasswordError(t *testing.T, gotErr, wantErr error) {
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
