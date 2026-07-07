package storage

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/account"
)

// AccountWriter implements account persistence using PostgreSQL via GORM.
type AccountWriter struct{}

// NewAccountWriter creates a new AccountWriter backed by the given GORM database connection.
func NewAccountWriter() *AccountWriter {
	return &AccountWriter{}
}

// Create persists a new Account to the database.
func (w *AccountWriter) Create(ctx context.Context, db *gorm.DB, acc *account.Account) error {
	if acc == nil {
		return account.ErrNilAccount
	}

	err := gorm.G[AccountRow](db).Create(ctx, accountToRow(acc))
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return account.ErrUsernameAlreadyExists
		}

		return fmt.Errorf("insert account: %w", err)
	}

	return nil
}

// UpdateLastLogin persists the account's last login time and IP to the database.
func (w *AccountWriter) UpdateLastLogin(ctx context.Context, db *gorm.DB, acc *account.Account) error {
	if acc == nil {
		return account.ErrNilAccount
	}

	accRow := accountToRow(acc)

	_, err := gorm.G[AccountRow](db).Where(AccountRowFields.ID.Eq(accRow.ID)).
		Set(AccountRowFields.LastLoginAt.Set(*accRow.LastLoginAt),
			AccountRowFields.LastLoginIP.Set(*accRow.LastLoginIP)).
		Update(ctx)
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}

	return nil
}
