package account_test

import (
	"errors"
	"fmt"
	"net/netip"
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

func TestAccountFields(t *testing.T) {
	t.Parallel()

	now := mustRegistrationTime()

	acc, err := account.NewAccount(validAccountID, usernameAlice, validPwd, now)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	if acc.ID != account.ID(validAccountID) {
		t.Errorf("ID = %d, want %d", acc.ID, validAccountID)
	}

	if acc.Username != account.Username(usernameAlice) {
		t.Errorf("Username = %q, want %q", acc.Username, usernameAlice)
	}

	if acc.DisplayName != usernameAlice {
		t.Errorf("DisplayName = %q, want %q", acc.DisplayName, usernameAlice)
	}

	if acc.Role != account.RoleMember {
		t.Errorf("Role = %v, want %v", acc.Role, account.RoleMember)
	}

	if acc.Status != account.StatusActive {
		t.Errorf("Status = %v, want %v", acc.Status, account.StatusActive)
	}

	if !acc.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", acc.CreatedAt, now)
	}

	if !acc.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt = %v, want %v", acc.UpdatedAt, now)
	}

	if acc.PasswordChangeRequired {
		t.Error("PasswordChangeRequired should be false")
	}

	if acc.LastLoginAt != nil {
		t.Errorf("LastLoginAt = %v, want nil", acc.LastLoginAt)
	}

	if acc.LastLoginIP != nil {
		t.Errorf("LastLoginIP = %v, want nil", acc.LastLoginIP)
	}

	if !acc.VerifyPassword(validPwd) {
		t.Error("VerifyPassword should match raw password")
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

	if acc.ID != validAccountID {
		t.Errorf("ID() = %d, want %d", acc.ID, validAccountID)
	}

	if acc.Username != usernameAlice {
		t.Errorf("Username() = %q, want %q", acc.Username, usernameAlice)
	}

	if acc.DisplayName != usernameAlice {
		t.Errorf("DisplayName() = %q, want %q", acc.DisplayName, usernameAlice)
	}
}

func assertAccountDefaults(t *testing.T, acc account.Account) {
	t.Helper()

	if acc.Role != account.RoleMember {
		t.Errorf("Role() = %v, want %v", acc.Role, account.RoleMember)
	}

	if acc.Status != account.StatusActive {
		t.Errorf("Status() = %v, want %v", acc.Status, account.StatusActive)
	}
}

func assertAccountAudit(t *testing.T, acc account.Account, now time.Time) {
	t.Helper()

	if !acc.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", acc.CreatedAt, now)
	}

	if !acc.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt = %v, want %v", acc.UpdatedAt, now)
	}

	if acc.LastLoginAt != nil {
		t.Errorf("LastLoginAt = %v, want nil", acc.LastLoginAt)
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

func TestAccountRecordLogin(t *testing.T) {
	t.Parallel()

	now := mustRegistrationTime()
	loginTime := now.Add(24 * time.Hour)

	acc, err := account.NewAccount(validAccountID, usernameAlice, validPwd, now)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	acc.RecordLogin(loginTime, netip.MustParseAddr("192.168.1.1"))

	if acc.LastLoginAt == nil {
		t.Fatal("LastLoginAt is nil after RecordLogin")
	}
	if !acc.LastLoginAt.Equal(loginTime) {
		t.Errorf("LastLoginAt = %v, want %v", acc.LastLoginAt, loginTime)
	}
	if acc.LastLoginIP == nil {
		t.Fatal("LastLoginIP is nil after RecordLogin")
	}
	if *acc.LastLoginIP != netip.MustParseAddr("192.168.1.1") {
		t.Errorf("LastLoginIP = %v, want %v", *acc.LastLoginIP, netip.MustParseAddr("192.168.1.1"))
	}
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
