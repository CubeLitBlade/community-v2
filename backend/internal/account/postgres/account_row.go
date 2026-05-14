package postgres

import "time"

type accountRow struct {
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

func (accountRow) TableName() string {
	return "accounts"
}
