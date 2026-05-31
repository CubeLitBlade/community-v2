package account

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// ByIDFinder is the interface for finding accounts by their unique ID.
type ByIDFinder interface {
	FindByID(ctx context.Context, db *gorm.DB, id int64) (*Account, error)
}

// Profile represents a public-facing account profile.
type Profile struct {
	ID          int64
	Username    string
	DisplayName string
}

// NewProfile creates a Profile from an Account domain object.
func NewProfile(account *Account) *Profile {
	return &Profile{
		ID:          account.ID(),
		Username:    account.Username(),
		DisplayName: account.DisplayName(),
	}
}

// ProfileFinder retrieves account profiles by ID.
type ProfileFinder struct {
	finder ByIDFinder
	db     *gorm.DB
	logger *slog.Logger
}

// NewProfileFinder creates a ProfileFinder backed by the given ByIDFinder.
func NewProfileFinder(finder ByIDFinder, db *gorm.DB, logger *slog.Logger) *ProfileFinder {
	if logger == nil {
		logger = slog.Default()
		logger.Warn("No logger provided while creating profile finder, using default logger.")
	}

	return &ProfileFinder{
		finder: finder,
		db:     db,
		logger: logger,
	}
}

// Find looks up an account by ID and returns its public Profile.
func (f *ProfileFinder) Find(ctx context.Context, id int64) (*Profile, error) {
	acc, err := f.finder.FindByID(ctx, f.db, id)
	if err != nil {
		f.logger.Error("Failed to find account", "id", id, "error", err)

		return nil, fmt.Errorf("find by id: %w", err)
	}

	return NewProfile(acc), nil
}
