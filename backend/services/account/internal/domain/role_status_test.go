package account_test

import (
	"testing"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
)

func TestParseRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    account.Role
		wantErr bool
	}{
		{"valid admin", account.RoleAdmin, account.RoleAdmin, false},
		{"valid moderator", account.RoleModerator, account.RoleModerator, false},
		{"valid member", account.RoleMember, account.RoleMember, false},
		{"invalid role", "superadmin", "", true},
		{"empty role", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := account.ParseRole(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseRole() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleIsValid(t *testing.T) {
	t.Parallel()

	validRoles := []account.Role{account.RoleAdmin, account.RoleModerator, account.RoleMember}
	for _, role := range validRoles {
		if !role.IsValid() {
			t.Errorf("%v.IsValid() = false, want true", role)
		}
	}

	if account.Role("bogus").IsValid() {
		t.Error("bogus role IsValid() = true, want false")
	}
}

func TestRoleString(t *testing.T) {
	t.Parallel()

	r := account.Role(account.RoleAdmin)
	if r.String() != account.RoleAdmin {
		t.Errorf("String() = %q, want %q", r.String(), account.RoleAdmin)
	}
}

func TestParseStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    account.Status
		wantErr bool
	}{
		{"valid active", "active", account.StatusActive, false},
		{"valid suspended", "suspended", account.StatusSuspended, false},
		{"valid archived", "archived", account.StatusArchived, false},
		{"invalid status", "deleted", "", true},
		{"empty status", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := account.ParseStatus(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseStatus() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusIsValid(t *testing.T) {
	t.Parallel()

	validStatuses := []account.Status{account.StatusActive, account.StatusSuspended, account.StatusArchived}
	for _, status := range validStatuses {
		if !status.IsValid() {
			t.Errorf("%v.IsValid() = false, want true", status)
		}
	}

	if account.Status("bogus").IsValid() {
		t.Error("bogus status IsValid() = true, want false")
	}
}

func TestStatusCanLogin(t *testing.T) {
	t.Parallel()

	if !account.StatusActive.CanLogin() {
		t.Error("StatusActive.CanLogin() = false, want true")
	}
	if account.StatusSuspended.CanLogin() {
		t.Error("StatusSuspended.CanLogin() = true, want false")
	}
	if account.StatusArchived.CanLogin() {
		t.Error("StatusArchived.CanLogin() = true, want false")
	}
}

func TestStatusCanBeSuspended(t *testing.T) {
	t.Parallel()

	if !account.StatusActive.CanBeSuspended() {
		t.Error("StatusActive.CanBeSuspended() = false, want true")
	}
	if account.StatusSuspended.CanBeSuspended() {
		t.Error("StatusSuspended.CanBeSuspended() = true, want false")
	}
}

func TestStatusCanBeRestored(t *testing.T) {
	t.Parallel()

	if !account.StatusSuspended.CanBeRestored() {
		t.Error("StatusSuspended.CanBeRestored() = false, want true")
	}
	if account.StatusActive.CanBeRestored() {
		t.Error("StatusActive.CanBeRestored() = true, want false")
	}
}

func TestStatusString(t *testing.T) {
	t.Parallel()

	if account.StatusActive.String() != "active" {
		t.Errorf("String() = %q, want %q", account.StatusActive.String(), "active")
	}
}

func TestMustParseRolePanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseRole should panic on invalid role")
		}
	}()

	_ = account.MustParseRole("__invalid__")
}
