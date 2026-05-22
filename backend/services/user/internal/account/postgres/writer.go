package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/cubelitblade/community-v2/backend/services/user/internal/account"
	"gorm.io/gorm"
)

// Writer implements account persistence using PostgreSQL via GORM.
type Writer struct {
	db *gorm.DB
}

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

// UpdateLastLogin persists the account's last login time and IP to the database.
func (w *Writer) UpdateLastLogin(ctx context.Context, acc *account.Account) error {
	if acc == nil {
		panic("nil account")
	}

	accRow := accountToRow(acc)

	_, err := gorm.G[Row](w.db).Where(RowFields.ID.Eq(accRow.ID)).
		Set(RowFields.LastLoginAt.Set(*accRow.LastLoginAt),
			RowFields.LastLoginIP.Set(*accRow.LastLoginIP)).
		Update(ctx)
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}

	return nil
}
