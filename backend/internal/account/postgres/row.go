package postgres

import "time"

// Row represents a record in the "accounts" PostgreSQL table.
// Pointer fields are used to represent nullable database columns
type Row struct {
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
func (Row) TableName() string {
	return "accounts"
}
