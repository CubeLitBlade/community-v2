package postgres

import (
	"context"
	"fmt"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	if db == nil {
		panic("postgres.NewRepository: db is nil")
	}

	return &Repository{
		db: db,
	}
}

var _ account.Repository = (*Repository)(nil)

func (r *Repository) Create(ctx context.Context, acc account.Account) error {
	po := toRow(acc)

	if err := r.db.WithContext(ctx).Create(&po).Error; err != nil {
		return fmt.Errorf("insert account: %w", err)
	}

	return nil
}

func toRow(acc account.Account) accountRow {
	s := acc.Snapshot()

	var lastLoginIP *string
	if s.LastLoginIP != nil {
		ip := s.LastLoginIP.String()
		lastLoginIP = &ip
	}

	return accountRow{
		ID:                     int64(s.ID),
		Username:               s.Username,
		PasswordHash:           s.PasswordHash,
		PasswordChangeRequired: s.PasswordChangeRequired,
		DisplayName:            s.DisplayName,
		Role:                   s.Role.String(),
		Status:                 s.Status.String(),
		CreatedAt:              s.CreatedAt,
		UpdatedAt:              s.UpdatedAt,
		LastLoginAt:            s.LastLoginAt,
		LastLoginIP:            lastLoginIP,
	}
}
