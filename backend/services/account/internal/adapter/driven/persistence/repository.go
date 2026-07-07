package persistence

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain/port"
)

var (
	_ port.AccountReader = (*AccountRepository)(nil)
	_ port.AccountWriter = (*AccountRepository)(nil)
)

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) FindByID(ctx context.Context, id int64) (*account.Account, error) {
	var row AccountRow
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrAccountNotFound
		}

		return nil, fmt.Errorf("find account by id: %w", err)
	}

	return rowToAccount(&row)
}

func (r *AccountRepository) FindByUsername(ctx context.Context, username account.Username) (*account.Account, error) {
	var row AccountRow
	err := r.db.WithContext(ctx).Where("username = ?", username.Value()).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrAccountNotFound
		}

		return nil, fmt.Errorf("find account by username: %w", err)
	}

	return rowToAccount(&row)
}

func (r *AccountRepository) Create(ctx context.Context, acc *account.Account) error {
	err := r.db.WithContext(ctx).Create(accountToRow(acc)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return account.ErrUsernameAlreadyExists
		}

		return fmt.Errorf("insert account: %w", err)
	}

	return nil
}

func (r *AccountRepository) Update(ctx context.Context, acc *account.Account) error {
	err := r.db.WithContext(ctx).Save(accountToRow(acc)).Error
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}

	return nil
}
