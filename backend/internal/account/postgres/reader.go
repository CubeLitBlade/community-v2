package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"gorm.io/gorm"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
)

// Reader implements account.Reader by querying a PostgreSQL database via Gorm.
type Reader struct {
	db *gorm.DB
}

var _ account.Reader = (*Reader)(nil)

// NewReader creates a new Reader backed by the given Gorm DB instance.
// It panics if db is nil.
func NewReader(db *gorm.DB) *Reader {
	if db == nil {
		panic("postgres.NewReader: db is nil")
	}

	return &Reader{
		db: db,
	}
}

// LookupByUsername retrieves the account that matches the given username.
// It returns account.ErrAccountNotFound if no matching record exists.
func (r *Reader) LookupByUsername(
	ctx context.Context,
	username string,
) (account.Account, error) {
	row, err := gorm.G[Row](r.db).
		Where("username = ?", username).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return account.Account{}, account.ErrAccountNotFound
		}

		return account.Account{}, fmt.Errorf(
			"lookup account by username: %w",
			err,
		)
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

// rowToAccount converts a database Row into a domain Account.
// If the row's LastLoginIP is non-nil and parseable, it is converted to a
// netip.Addr; otherwise LastLoginIP on the resulting account is left nil.
func rowToAccount(row *Row) account.Account {
	var lastLoginIP *netip.Addr

	if row.LastLoginIP != nil {
		ip, err := netip.ParseAddr(*row.LastLoginIP)
		if err == nil {
			lastLoginIP = new(ip)
		}
	}

	snap := account.Snapshot{
		ID:                     row.ID,
		Username:               row.Username,
		PasswordHash:           row.PasswordHash,
		PasswordChangeRequired: row.PasswordChangeRequired,
		DisplayName:            row.DisplayName,
		Role:                   row.Role,
		Status:                 row.Status,
		CreatedAt:              row.CreatedAt,
		UpdatedAt:              row.UpdatedAt,
		LastLoginAt:            row.LastLoginAt,
		LastLoginIP:            lastLoginIP,
	}

	return account.NewAccountFromSnapshot(snap)
}
