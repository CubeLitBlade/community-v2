package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
)

// Writer implements account persistence using PostgreSQL via GORM.
type Writer struct {
	db *gorm.DB
}

var _ account.Creator = (*Writer)(nil)

// NewWriter creates a new Writer backed by the given GORM database
// connection.
// Panics if db is nil.
func NewWriter(db *gorm.DB) *Writer {
	if db == nil {
		panic("nil db")
	}

	return &Writer{
		db: db,
	}
}

// Create persists a new Account to the database.
func (w *Writer) Create(ctx context.Context, acc *account.Account) error {
	if acc == nil {
		panic("nil account")
	}

	err := gorm.G[Row](w.db).Create(ctx, accountToRow(acc))
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return account.ErrUsernameAlreadyExists
		}

		return fmt.Errorf("insert account: %w", err)
	}

	return nil
}

func accountToRow(acc *account.Account) *Row {
	snapshot := acc.Snapshot()

	var lastLoginIP *string

	if snapshot.LastLoginIP != nil {
		lastLoginIP = new(snapshot.LastLoginIP.String())
	}

	return &Row{
		ID:                     snapshot.ID,
		Username:               snapshot.Username,
		PasswordHash:           snapshot.PasswordHash,
		PasswordChangeRequired: snapshot.PasswordChangeRequired,
		DisplayName:            snapshot.DisplayName,
		Role:                   snapshot.Role,
		Status:                 snapshot.Status,
		CreatedAt:              snapshot.CreatedAt,
		UpdatedAt:              snapshot.UpdatedAt,
		LastLoginAt:            snapshot.LastLoginAt,
		LastLoginIP:            lastLoginIP,
	}
}
