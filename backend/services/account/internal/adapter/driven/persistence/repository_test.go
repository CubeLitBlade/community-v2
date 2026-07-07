package persistence_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/adapter/driven/persistence"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
)

func TestAccountRepository_CreateAndFindByID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := persistence.NewAccountRepository(db)
	ctx := context.Background()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	acc, err := account.NewAccount(1, "Alice", "this-is-a-valid-password", now)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	if err := repo.Create(ctx, &acc); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.FindByID(ctx, 1)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.ID != account.ID(1) {
		t.Errorf("ID = %d, want 1", found.ID)
	}
	if found.Username != account.Username("Alice") {
		t.Errorf("Username = %q, want Alice", found.Username)
	}
	if !found.VerifyPassword("this-is-a-valid-password") {
		t.Error("VerifyPassword should return true")
	}
	if found.Role != account.RoleMember {
		t.Errorf("Role = %v, want member", found.Role)
	}
	if found.Status != account.StatusActive {
		t.Errorf("Status = %v, want active", found.Status)
	}
	if found.DisplayName != "Alice" {
		t.Errorf("DisplayName = %q, want Alice", found.DisplayName)
	}
}

func TestAccountRepository_FindByUsername(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := persistence.NewAccountRepository(db)
	ctx := context.Background()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	acc, err := account.NewAccount(2, "Bob", "this-is-a-valid-password", now)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	if err := repo.Create(ctx, &acc); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	username, err := account.NewUsername("Bob")
	if err != nil {
		t.Fatalf("NewUsername() error = %v", err)
	}

	found, err := repo.FindByUsername(ctx, username)
	if err != nil {
		t.Fatalf("FindByUsername() error = %v", err)
	}

	if found.ID != account.ID(2) {
		t.Errorf("ID = %d, want 2", found.ID)
	}
}

func TestAccountRepository_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := persistence.NewAccountRepository(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, 999)
	if !errors.Is(err, account.ErrAccountNotFound) {
		t.Errorf("error = %v, want %v", err, account.ErrAccountNotFound)
	}
}

func TestAccountRepository_FindByUsername_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := persistence.NewAccountRepository(db)
	ctx := context.Background()

	username, err := account.NewUsername("Nobody")
	if err != nil {
		t.Fatalf("NewUsername() error = %v", err)
	}

	_, err = repo.FindByUsername(ctx, username)
	if !errors.Is(err, account.ErrAccountNotFound) {
		t.Errorf("error = %v, want %v", err, account.ErrAccountNotFound)
	}
}

func TestAccountRepository_Create_DuplicateUsername(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := persistence.NewAccountRepository(db)
	ctx := context.Background()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	acc1, err := account.NewAccount(3, "Charlie", "this-is-a-valid-password", now)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	if err := repo.Create(ctx, &acc1); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	acc2, err := account.NewAccount(4, "Charlie", "this-is-a-valid-password", now)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	err = repo.Create(ctx, &acc2)
	if !errors.Is(err, account.ErrUsernameAlreadyExists) {
		t.Errorf("error = %v, want %v", err, account.ErrUsernameAlreadyExists)
	}
}

func TestAccountRepository_Update(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := persistence.NewAccountRepository(db)
	ctx := context.Background()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	acc, err := account.NewAccount(5, "Dave", "this-is-a-valid-password", now)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	if err := repo.Create(ctx, &acc); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	loginTime := now.Add(time.Hour)
	ip := netip.MustParseAddr("10.0.0.1")
	acc.RecordLogin(loginTime, ip)

	if err := repo.Update(ctx, &acc); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := repo.FindByID(ctx, 5)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if updated.LastLoginAt == nil || !updated.LastLoginAt.Equal(loginTime) {
		t.Errorf("LastLoginAt = %v, want %v", updated.LastLoginAt, loginTime)
	}
	if updated.LastLoginIP == nil || *updated.LastLoginIP != ip {
		t.Errorf("LastLoginIP = %v, want %v", updated.LastLoginIP, ip)
	}
}

func TestAccountRepository_CreateWithLastLogin(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := persistence.NewAccountRepository(db)
	ctx := context.Background()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	acc, err := account.NewAccount(6, "Eve", "this-is-a-valid-password", now)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	loginTime := now.Add(2 * time.Hour)
	ip := netip.MustParseAddr("192.168.1.1")
	acc.RecordLogin(loginTime, ip)

	if err := repo.Create(ctx, &acc); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.FindByID(ctx, 6)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.LastLoginAt == nil || !found.LastLoginAt.Equal(loginTime) {
		t.Errorf("LastLoginAt = %v, want %v", found.LastLoginAt, loginTime)
	}
	if found.LastLoginIP == nil || *found.LastLoginIP != ip {
		t.Errorf("LastLoginIP = %v, want %v", found.LastLoginIP, ip)
	}
}
