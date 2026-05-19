package account

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"time"
)

type LastLoginUpdater interface {
	UpdateLastLogin(ctx context.Context, acc *Account) error
}

type LoginRecorder struct {
	now     func() time.Time
	updater LastLoginUpdater
	logger  *slog.Logger
}

func NewLoginRecorder(
	now func() time.Time, updater LastLoginUpdater, logger *slog.Logger,
) *LoginRecorder {
	return &LoginRecorder{
		now:     now,
		updater: updater,
		logger:  logger,
	}
}

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
