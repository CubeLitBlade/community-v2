package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
)

var errAccountNil = errors.New("account is nil")

// Repository implements account persistence using PostgreSQL via GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new Repository backed by the given GORM database
// connection.
// Panics if db is nil.
func NewRepository(db *gorm.DB) *Repository {
	if db == nil {
		panic("postgres.NewRepository: db is nil")
	}

	return &Repository{
		db: db,
	}
}

// Create persists a new Account to the database.
func (r *Repository) Create(ctx context.Context, acc *account.Account) error {
	if acc == nil {
		return errAccountNil
	}

	po := toRow(acc)

	err := r.db.WithContext(ctx).Create(&po).Error
	if err != nil {
		return fmt.Errorf("insert account: %w", err)
	}

	return nil
}

func toRow(acc *account.Account) accountRow {
	snapshot := acc.Snapshot()

	var lastLoginIP *string

	if snapshot.LastLoginIP != nil {
		ip := snapshot.LastLoginIP.String()
		lastLoginIP = &ip
	}

	return accountRow{
		ID:                     int64(snapshot.ID),
		Username:               snapshot.Username,
		PasswordHash:           snapshot.PasswordHash,
		PasswordChangeRequired: snapshot.PasswordChangeRequired,
		DisplayName:            snapshot.DisplayName,
		Role:                   snapshot.Role.String(),
		Status:                 snapshot.Status.String(),
		CreatedAt:              snapshot.CreatedAt,
		UpdatedAt:              snapshot.UpdatedAt,
		LastLoginAt:            snapshot.LastLoginAt,
		LastLoginIP:            lastLoginIP,
	}
}
