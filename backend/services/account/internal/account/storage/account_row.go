package storage

import (
	"net/netip"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/account"
	"gorm.io/cli/gorm/field"
)

// AccountRow represents a record in the "accounts" PostgreSQL table.
// Pointer fields are used to represent nullable database columns
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

// TableName overrides the default Gorm table name to "accounts".
func (AccountRow) TableName() string {
	return "accounts"
}

// AccountRowFields provides strongly-typed Gorm field references for the accounts table columns.
var AccountRowFields = struct {
	CreatedAt              field.Time
	UpdatedAt              field.Time
	LastLoginAt            field.Time
	LastLoginIP            field.String
	Username               field.String
	PasswordHash           field.String
	DisplayName            field.String
	Role                   field.String
	Status                 field.String
	ID                     field.Number[int64]
	PasswordChangeRequired field.Bool
}{
	CreatedAt:              field.Time{}.WithColumn("created_at"),
	UpdatedAt:              field.Time{}.WithColumn("updated_at"),
	LastLoginAt:            field.Time{}.WithColumn("last_login_at"),
	LastLoginIP:            field.String{}.WithColumn("last_login_ip"),
	Username:               field.String{}.WithColumn("username"),
	PasswordHash:           field.String{}.WithColumn("password_hash"),
	DisplayName:            field.String{}.WithColumn("display_name"),
	Role:                   field.String{}.WithColumn("role"),
	Status:                 field.String{}.WithColumn("status"),
	ID:                     field.Number[int64]{}.WithColumn("id"),
	PasswordChangeRequired: field.Bool{}.WithColumn("password_change_required"),
}

// accountToRow converts a domain Account into a database AccountRow for persistence.
func accountToRow(acc *account.Account) *AccountRow {
	snapshot := acc.Snapshot()

	var lastLoginIP *string

	if snapshot.LastLoginIP != nil {
		lastLoginIP = new(snapshot.LastLoginIP.String())
	}

	return &AccountRow{
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

// rowToAccount converts a database AccountRow into a domain Account.
// If the row's LastLoginIP is non-nil and parseable, it is converted to a
// netip.Addr; otherwise LastLoginIP on the resulting account is left nil.
func rowToAccount(row *AccountRow) *account.Account {
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

	return new(account.NewAccountFromSnapshot(snap))
}
