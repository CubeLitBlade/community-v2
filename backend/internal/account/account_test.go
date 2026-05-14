//nolint:testpackage // White-box domain tests inspect unexported invariants without exposing getters only for tests.
package account

import (
	"errors"
	"testing"
	"time"
)

const (
	usernameAlice  = "alice"
	validAccountID = ID(1)
	validPassword  = "this-is-a-valid-password"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	now := registrationTime()

	acc, err := Register(ID(1), usernameAlice, validPassword, now)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	assertAccountIdentity(t, acc)
	assertAccountDefaults(t, acc)
	assertAccountAudit(t, acc, now)
	assertPasswordHashMatches(t, acc.passwordHash, validPassword)
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	now := registrationTime()

	for _, testCase := range invalidRegisterCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := Register(testCase.id, testCase.username, testCase.password, now)

			if !errors.Is(err, testCase.wantError) {
				t.Fatalf("Register() error = %v, want %v", err, testCase.wantError)
			}
		})
	}
}

func registrationTime() time.Time {
	return time.Date(2026, 5, 13, 1, 30, 0, 0, time.UTC)
}

func assertAccountIdentity(t *testing.T, acc Account) {
	t.Helper()

	if acc.id != ID(1) {
		t.Errorf("id = %d, want %d", acc.id, ID(1))
	}

	if acc.username.Value() != usernameAlice {
		t.Errorf("username = %q, want %q", acc.username.Value(), "alice")
	}

	if acc.displayName != "alice" {
		t.Errorf("displayName = %q, want %q", acc.displayName, "alice")
	}
}

func assertAccountDefaults(t *testing.T, acc Account) {
	t.Helper()

	if acc.passwordChangeRequired {
		t.Errorf("passwordChangeRequired = true, want false")
	}

	if acc.role != RoleMember {
		t.Errorf("role = %v, want %v", acc.role, RoleMember)
	}

	if acc.status != StatusActive {
		t.Errorf("status = %v, want %v", acc.status, StatusActive)
	}
}

func assertAccountAudit(t *testing.T, acc Account, now time.Time) {
	t.Helper()

	if !acc.audit.createdAt.Equal(now) {
		t.Errorf("createdAt = %v, want %v", acc.audit.createdAt, now)
	}

	if !acc.audit.updatedAt.Equal(now) {
		t.Errorf("updatedAt = %v, want %v", acc.audit.updatedAt, now)
	}

	if acc.audit.lastLogin != nil {
		t.Errorf("lastLogin = %v, want nil", acc.audit.lastLogin)
	}
}

func assertPasswordHashMatches(t *testing.T, hash PasswordHash, raw string) {
	t.Helper()

	if !PasswordMatches(hash, raw) {
		t.Errorf("passwordHash should match raw password")
	}
}

type invalidRegisterCase struct {
	wantError error
	name      string
	username  string
	password  string
	id        ID
}

func invalidRegisterCases() []invalidRegisterCase {
	return []invalidRegisterCase{
		{
			name:      "zero id",
			id:        0,
			username:  usernameAlice,
			password:  validPassword,
			wantError: ErrInvalidAccountID,
		},
		{
			name:      "negative id",
			id:        -1,
			username:  usernameAlice,
			password:  validPassword,
			wantError: ErrInvalidAccountID,
		},
		{
			name:      "empty username",
			id:        validAccountID,
			username:  "",
			password:  validPassword,
			wantError: ErrUsernameBlank,
		},
		{
			name:      "short password",
			id:        validAccountID,
			username:  usernameAlice,
			password:  "short",
			wantError: ErrPasswordTooShort,
		},
	}
}
