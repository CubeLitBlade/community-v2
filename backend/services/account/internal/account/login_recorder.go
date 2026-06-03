package account

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"time"

	"gorm.io/gorm"
)

// LoginRecorderOption defines a functional option for configuring an LoginRecorder.
type LoginRecorderOption func(*LoginRecorder)

// LastLoginUpdater records the last login timestamp for an account.
type LastLoginUpdater interface {
	UpdateLastLogin(ctx context.Context, db *gorm.DB, acc *Account) error
}

// LoginRecorder records the last login details after a user logs in.
// It looks up the account by ID, updates its last login info, then persists the change via the injected updater.
type LoginRecorder struct {
	finder  ByIDFinder
	updater LastLoginUpdater
	db      *gorm.DB
	clock   func() time.Time
	logger  *slog.Logger
}

// NewLoginRecorder creates a new LoginRecorder.
func NewLoginRecorder(finder ByIDFinder, updater LastLoginUpdater,
	db *gorm.DB, logger *slog.Logger, opts ...LoginRecorderOption,
) *LoginRecorder {
	logger = logger.With(
		slog.String("service", "account"),
		slog.String("component", "account/login_recorder"),
	)

	loginRecorder := &LoginRecorder{
		finder:  finder,
		updater: updater,
		db:      db,
		clock:   time.Now,
		logger:  logger,
	}

	for _, opt := range opts {
		opt(loginRecorder)
	}

	return loginRecorder
}

// Record records a login event for the given account ID.
// It looks up the account, updates its last login time and IP, then persists the change.
func (r *LoginRecorder) Record(ctx context.Context, accountID int64, ipaddr netip.Addr) error {
	acc, err := r.finder.FindByID(ctx, r.db, accountID)
	if err != nil {
		return fmt.Errorf("find by id: %w", err)
	}

	acc.RecordLogin(r.clock(), ipaddr)

	if err := r.updater.UpdateLastLogin(ctx, r.db, acc); err != nil {
		return fmt.Errorf("update last login: %w", err)
	}

	return nil
}

// WithLoginRecorderClock sets the clock function for the LoginRecorder.
func WithLoginRecorderClock(clock func() time.Time) LoginRecorderOption {
	return func(a *LoginRecorder) {
		a.clock = clock
	}
}
