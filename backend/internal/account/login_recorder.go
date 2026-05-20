package account

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"time"
)

type lastLoginUpdater interface {
	UpdateLastLogin(ctx context.Context, acc *Account) error
}

// LoginRecorder records the last login details after a user logs in.
// It uses an injected lastLoginUpdater to persist the updated account information.
type LoginRecorder struct {
	now     func() time.Time
	updater lastLoginUpdater
	logger  *slog.Logger
}

// NewLoginRecorder creates a new LoginRecorder.
// now provides the current time;
// updater persists the account's last login info;
// logger logs operational events.
func NewLoginRecorder(
	now func() time.Time, updater lastLoginUpdater, logger *slog.Logger,
) *LoginRecorder {
	return &LoginRecorder{
		now:     now,
		updater: updater,
		logger:  logger,
	}
}

// Record records a login event.
// It updates the account's last login time and IP,
// then persists the change via the updater.
func (r *LoginRecorder) Record(
	ctx context.Context, acc *Account, ip netip.Addr,
) error {
	acc.RecordLogin(r.now(), ip)
	err := r.updater.UpdateLastLogin(ctx, acc)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}

	return nil
}
