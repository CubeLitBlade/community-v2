package account

import (
	"errors"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	now := time.Date(2026, 5, 13, 1, 30, 0, 0, time.UTC)

	acc, err := Register(1, "alice", "this-is-a-valid-password", now)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if acc.id != 1 {
		t.Errorf("id = %d, want %d", acc.id, ID(1))
	}

	if acc.username.Value() != "alice" {
		t.Errorf("username = %q, want %q", acc.username.Value(), "alice")
	}

	if acc.displayName != "alice" {
		t.Errorf("displayName = %q, want %q", acc.displayName, "alice")
	}

	if acc.passwordChangeRequired {
		t.Errorf("passwordChangeRequired = true, want false")
	}

	if acc.role != RoleMember {
		t.Errorf("role = %v, want %v", acc.role, RoleMember)
	}

	if acc.status != StatusActive {
		t.Errorf("status = %v, want %v", acc.status, StatusActive)
	}

	if !acc.audit.createdAt.Equal(now) {
		t.Errorf("createdAt = %v, want %v", acc.audit.createdAt, now)
	}

	if !acc.audit.updatedAt.Equal(now) {
		t.Errorf("updatedAt = %v, want %v", acc.audit.updatedAt, now)
	}

	if acc.audit.lastLogin != nil {
		t.Errorf("lastLoginAt = %v, want nil", acc.audit.lastLogin.at)
		t.Errorf("lastLoginIP = %v, want nil", acc.audit.lastLogin.ip)
	}

	if !PasswordMatches(acc.passwordHash, "this-is-a-valid-password") {
		t.Errorf("passwordHash should match raw password")
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	now := time.Date(2026, 5, 13, 1, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		id        ID
		username  string
		password  string
		wantError error
	}{
		{
			name:      "zero id",
			id:        0,
			username:  "alice",
			password:  "this-is-a-valid-password",
			wantError: ErrInvalidAccountID,
		},
		{
			name:      "negative id",
			id:        -1,
			username:  "alice",
			password:  "this-is-a-valid-password",
			wantError: ErrInvalidAccountID,
		},
		{
			name:      "empty username",
			id:        1,
			username:  "",
			password:  "this-is-a-valid-password",
			wantError: ErrUsernameBlank,
		},
		{
			name:      "short password",
			id:        1,
			username:  "alice",
			password:  "short",
			wantError: ErrPasswordTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Register(tt.id, tt.username, tt.password, now)

			if !errors.Is(err, tt.wantError) {
				t.Fatalf("Register() error = %v, want %v", err, tt.wantError)
			}
		})
	}
}
