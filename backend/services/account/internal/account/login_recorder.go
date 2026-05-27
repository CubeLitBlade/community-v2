package account

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"time"
)

// LastLoginUpdater records the last login timestamp for an account.
type LastLoginUpdater interface {
	UpdateLastLogin(ctx context.Context, acc *Account) error
}

// LoginRecorder records the last login details after a user logs in.
// It looks up the account by ID, updates its last login info,
// then persists the change via the injected updater.
type LoginRecorder struct {
	now     func() time.Time
	finder  ByIDFinder
	updater LastLoginUpdater
	logger  *slog.Logger
}

// NewLoginRecorder creates a new LoginRecorder.
// now provides the current time; finder looks up accounts by ID;
// updater persists the account's last login info;
// logger logs operational events.
func NewLoginRecorder(
	now func() time.Time, finder ByIDFinder, updater LastLoginUpdater, logger *slog.Logger,
) *LoginRecorder {
	return &LoginRecorder{
		now:     now,
		finder:  finder,
		updater: updater,
		logger:  logger,
	}
}

// Record records a login event for the given account ID.
// It looks up the account, updates its last login time and IP,
// then persists the change.
func (r *LoginRecorder) Record(
	ctx context.Context, accountID int64, ipaddr netip.Addr,
) error {
	acc, err := r.finder.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("find by id: %w", err)
	}

	acc.RecordLogin(r.now(), ipaddr)
	if err := r.updater.UpdateLastLogin(ctx, acc); err != nil {
		return fmt.Errorf("update last login: %w", err)
	}

	return nil
}
