package persistence

import (
	"net/netip"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
)

type AccountRow struct {
	CreatedAt              time.Time  `gorm:"column:created_at"`
	UpdatedAt              time.Time  `gorm:"column:updated_at"`
	LastLoginAt            *time.Time `gorm:"column:last_login_at"`
	LastLoginIP            *string    `gorm:"column:last_login_ip"`
	Username               string     `gorm:"column:username"`
	PasswordHash           string     `gorm:"column:password_hash"`
	DisplayName            string     `gorm:"column:display_name"`
	Role                   string     `gorm:"column:role"`
	Status                 string     `gorm:"column:status"`
	ID                     int64      `gorm:"column:id;primaryKey"`
	PasswordChangeRequired bool       `gorm:"column:password_change_required"`
}

func (AccountRow) TableName() string {
	return "accounts"
}

func accountToRow(acc *account.Account) *AccountRow {
	var lastLoginIP *string
	if acc.LastLoginIP != nil {
		s := acc.LastLoginIP.String()
		lastLoginIP = &s
	}

	return &AccountRow{
		ID:                     int64(acc.ID),
		Username:               acc.Username.Value(),
		PasswordHash:           acc.PasswordHash.Value(),
		PasswordChangeRequired: acc.PasswordChangeRequired,
		DisplayName:            acc.DisplayName,
		Role:                   acc.Role.String(),
		Status:                 acc.Status.String(),
		CreatedAt:              acc.CreatedAt,
		UpdatedAt:              acc.UpdatedAt,
		LastLoginAt:            acc.LastLoginAt,
		LastLoginIP:            lastLoginIP,
	}
}

func rowToAccount(row *AccountRow) (*account.Account, error) {
	username, err := account.UsernameFromStorage(row.Username)
	if err != nil {
		return nil, err
	}

	passwordHash, err := account.PasswordHashFromStorage(row.PasswordHash)
	if err != nil {
		return nil, err
	}

	var lastLoginIP *netip.Addr
	if row.LastLoginIP != nil {
		ip, parseErr := netip.ParseAddr(*row.LastLoginIP)
		if parseErr == nil {
			lastLoginIP = &ip
		}
	}

	return &account.Account{
		ID:                     account.ID(row.ID),
		Username:               username,
		PasswordHash:           passwordHash,
		PasswordChangeRequired: row.PasswordChangeRequired,
		DisplayName:            row.DisplayName,
		Role:                   account.MustParseRole(row.Role),
		Status:                 account.MustParseStatus(row.Status),
		CreatedAt:              row.CreatedAt,
		UpdatedAt:              row.UpdatedAt,
		LastLoginAt:            row.LastLoginAt,
		LastLoginIP:            lastLoginIP,
	}, nil
}
