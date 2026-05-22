package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/cubelitblade/community-v2/backend/services/user/internal/account"
	"gorm.io/gorm"
)

// Reader implements account.Reader by querying a PostgreSQL database via Gorm.
type Reader struct {
	db *gorm.DB
}

// NewReader creates a new Reader backed by the given Gorm DB instance.
// It panics if db is nil.
func NewReader(db *gorm.DB) *Reader {
	if db == nil {
		panic("nil db")
	}

	return &Reader{
		db: db,
	}
}

// FindByUsername retrieves the account that matches the given username.
// It returns account.ErrAccountNotFound if no matching record exists.
func (r *Reader) FindByUsername(
	ctx context.Context,
	username string,
) (*account.Account, error) {
	row, err := gorm.G[Row](r.db).
		Where("username = ?", username).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrAccountNotFound
		}

		return nil, fmt.Errorf("find account by username: %w", err)
	}

	acc := rowToAccount(&row)

	return acc, nil
}

// FindByID retrieves the account that matches the given ID.
// It returns account.ErrAccountNotFound if no matching record exists.
func (r *Reader) FindByID(
	ctx context.Context,
	id int64,
) (*account.Account, error) {
	row, err := gorm.G[Row](r.db).Where(RowFields.ID.Eq(id)).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrAccountNotFound
		}

		return nil, fmt.Errorf("find account by id: %w", err)
	}

	acc := rowToAccount(&row)

	return acc, nil
}

// ExistsUsername reports whether an account with the given username already
// exists.
func (r *Reader) ExistsUsername(
	ctx context.Context,
	username string,
) (bool, error) {
	var exists int

	err := gorm.G[Row](r.db).
		Select("1").
		Where("username = ?", username).
		Limit(1).
		Scan(ctx, &exists)
	if err != nil {
		return false, fmt.Errorf(
			"check if username exists: %w",
			err,
		)
	}

	return exists > 0, nil
}
