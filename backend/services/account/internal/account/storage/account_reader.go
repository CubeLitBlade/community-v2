package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/account"
	"gorm.io/gorm"
)

// AccountReader implements account.Reader by querying a PostgreSQL database via Gorm.
type AccountReader struct{}

// NewAccountReader creates a new AccountReader backed by the given Gorm DB instance.
// It panics if db is nil.
func NewAccountReader() *AccountReader {
	return &AccountReader{}
}

// FindByUsername retrieves the account that matches the given username.
// It returns account.ErrAccountNotFound if no matching record exists.
func (r *AccountReader) FindByUsername(
	ctx context.Context, db *gorm.DB, username string,
) (*account.Account, error) {
	row, err := gorm.G[AccountRow](db).Where(AccountRowFields.Username.Eq(username)).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrAccountNotFound
		}

		return nil, fmt.Errorf("find account by username: %w", err)
	}

	return rowToAccount(&row), nil
}

// FindByID retrieves the account that matches the given ID.
// It returns account.ErrAccountNotFound if no matching record exists.
func (r *AccountReader) FindByID(
	ctx context.Context, db *gorm.DB, id int64,
) (*account.Account, error) {
	row, err := gorm.G[AccountRow](db).Where(AccountRowFields.ID.Eq(id)).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrAccountNotFound
		}

		return nil, fmt.Errorf("find account by id: %w", err)
	}

	acc := rowToAccount(&row)

	return acc, nil
}
