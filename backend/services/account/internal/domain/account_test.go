package account_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
)

const (
	usernameAlice       = "Alice"
	validAccountID      = 1
	zeroAccountID       = 0
	negativeAccountID   = -1
	registrationTimeRaw = "2026-05-13T00:00:00Z"
)

func TestNewAccount(t *testing.T) {
	t.Parallel()

	now := mustRegistrationTime()

	acc, err := account.NewAccount(validAccountID, usernameAlice, validPwd, now)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	assertAccountIdentity(t, acc)
	assertAccountDefaults(t, acc)
	assertAccountAudit(t, acc, now)
	assertPasswordMatches(t, acc, validPwd)
}

func TestNewAccountRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	now := mustRegistrationTime()

	for _, testCase := range invalidRegisterCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := account.NewAccount(
				testCase.id,
				testCase.username,
				testCase.password,
				now,
			)

			if !errors.Is(err, testCase.wantError) {
				t.Fatalf(
					"NewAccount() error = %v, want %v",
					err,
					testCase.wantError,
				)
			}
		})
	}
}

func TestAccountVerifyPassword(t *testing.T) {
	t.Parallel()

	now := mustRegistrationTime()

	acc, err := account.NewAccount(validAccountID, usernameAlice, validPwd, now)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	if !acc.VerifyPassword(validPwd) {
		t.Error("VerifyPassword() should return true for correct password")
	}

	if acc.VerifyPassword("wrong") {
		t.Error("VerifyPassword() should return false for incorrect password")
	}
}

func TestAccountSnapshotRoundtrip(t *testing.T) {
	t.Parallel()

	now := mustRegistrationTime()

	acc, err := account.NewAccount(validAccountID, usernameAlice, validPwd, now)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	snap := acc.Snapshot()
	restored := account.NewAccountFromSnapshot(snap)

	if restored.ID() != acc.ID() {
		t.Errorf("ID mismatch: %d != %d", restored.ID(), acc.ID())
	}

	if restored.Username() != acc.Username() {
		t.Errorf("Username mismatch: %q != %q", restored.Username(), acc.Username())
	}

	if restored.Role() != acc.Role() {
		t.Errorf("Role mismatch: %v != %v", restored.Role(), acc.Role())
	}

	if !restored.VerifyPassword(validPwd) {
		t.Error("restored account should verify correct password")
	}
}

func mustRegistrationTime() time.Time {
	t, err := time.Parse(time.RFC3339, registrationTimeRaw)
	if err != nil {
		panic(fmt.Sprintf("invalid registration time constant: %v", err))
	}

	return t
}

func assertAccountIdentity(t *testing.T, acc account.Account) {
	t.Helper()

	if acc.ID() != validAccountID {
		t.Errorf("ID() = %d, want %d", acc.ID(), validAccountID)
	}

	if acc.Username() != usernameAlice {
		t.Errorf("Username() = %q, want %q", acc.Username(), usernameAlice)
	}

	if acc.DisplayName() != usernameAlice {
		t.Errorf("DisplayName() = %q, want %q", acc.DisplayName(), usernameAlice)
	}
}

func assertAccountDefaults(t *testing.T, acc account.Account) {
	t.Helper()

	if acc.Role() != account.RoleMember {
		t.Errorf("Role() = %v, want %v", acc.Role(), account.RoleMember)
	}

	if acc.Status() != account.StatusActive {
		t.Errorf("Status() = %v, want %v", acc.Status(), account.StatusActive)
	}
}

func assertAccountAudit(t *testing.T, acc account.Account, now time.Time) {
	t.Helper()

	snap := acc.Snapshot()

	if !snap.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", snap.CreatedAt, now)
	}

	if !snap.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt = %v, want %v", snap.UpdatedAt, now)
	}

	if snap.LastLoginAt != nil {
		t.Errorf("LastLoginAt = %v, want nil", snap.LastLoginAt)
	}
}

func assertPasswordMatches(t *testing.T, acc account.Account, raw string) {
	t.Helper()

	if !acc.VerifyPassword(raw) {
		t.Error("VerifyPassword should match raw password")
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
			password:  validPwd,
			wantError: account.ErrInvalidAccountID,
		},
		{
			name:      "negative id",
			id:        negativeAccountID,
			username:  usernameAlice,
			password:  validPwd,
			wantError: account.ErrInvalidAccountID,
		},
		{
			name:      "empty username",
			id:        validAccountID,
			username:  "",
			password:  validPwd,
			wantError: account.ErrUsernameBlank,
		},
		{
			name:      "short password",
			id:        validAccountID,
			username:  usernameAlice,
			password:  "short",
			wantError: account.ErrPasswordTooShort,
		},
	}
}
