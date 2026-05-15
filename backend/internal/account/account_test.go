//nolint:testpackage // White-box domain tests
package account

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

const (
	usernameAlice       = "alice"
	validAccountID      = 1
	zeroAccountID       = 0
	negativeAccountID   = -1
	validPassword       = "this-is-a-valid-password"
	registrationTimeRaw = "2026-05-13T00:00:00Z"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	now := mustRegistrationTime()

	acc, err := Register(validAccountID, usernameAlice, validPassword, now)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	assertAccountIdentity(t, &acc)
	assertAccountDefaults(t, &acc)
	assertAccountAudit(t, &acc, now)
	assertPasswordHashMatches(t, acc.passwordHash, validPassword)
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	now := mustRegistrationTime()

	for _, testCase := range invalidRegisterCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := Register(
				testCase.id,
				testCase.username,
				testCase.password,
				now,
			)

			if !errors.Is(err, testCase.wantError) {
				t.Fatalf(
					"Register() error = %v, want %v",
					err,
					testCase.wantError,
				)
			}
		})
	}
}

func mustRegistrationTime() time.Time {
	t, err := time.Parse(time.RFC3339, registrationTimeRaw)
	if err != nil {
		panic(fmt.Sprintf("invalid registration time constant: %v", err))
	}

	return t
}

func assertAccountIdentity(t *testing.T, acc *Account) {
	t.Helper()

	if acc.id != ID(validAccountID) {
		t.Errorf("id = %d, want %d", acc.id, ID(validAccountID))
	}

	if acc.username.Value() != usernameAlice {
		t.Errorf("username = %q, want %q", acc.username.Value(), usernameAlice)
	}

	if acc.displayName != usernameAlice {
		t.Errorf("displayName = %q, want %q", acc.displayName, usernameAlice)
	}
}

func assertAccountDefaults(t *testing.T, acc *Account) {
	t.Helper()

	if acc.passwordChangeRequired {
		t.Error("passwordChangeRequired = true, want false")
	}

	if acc.role != RoleMember {
		t.Errorf("role = %v, want %v", acc.role, RoleMember)
	}

	if acc.status != StatusActive {
		t.Errorf("status = %v, want %v", acc.status, StatusActive)
	}
}

func assertAccountAudit(t *testing.T, acc *Account, now time.Time) {
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
		t.Error("passwordHash should match raw password")
	}
}

type invalidRegisterCase struct {
	wantError error
	name      string
	username  string
	password  string
	id        int64
}

func invalidRegisterCases() []invalidRegisterCase {
	return []invalidRegisterCase{
		{
			name:      "zero id",
			id:        zeroAccountID,
			username:  usernameAlice,
			password:  validPassword,
			wantError: ErrInvalidAccountID,
		},
		{
			name:      "negative id",
			id:        negativeAccountID,
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
