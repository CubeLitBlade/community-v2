package account

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/CubeLitBlade/community-v2/backend/pkg/common/idgen"
)

type creator interface {
	Create(ctx context.Context, acc *Account) error
}

// Registrar orchestrates the account registration process.
type Registrar struct {
	ids     idgen.Generator
	creator creator
	now     func() time.Time
	logger  *slog.Logger
}

// NewRegistrar creates and returns a new Registrar.
func NewRegistrar(
	ids idgen.Generator,
	creator creator,
	logger *slog.Logger,
) *Registrar {
	if logger == nil {
		panic("nil logger")
	}

	return &Registrar{
		ids:     ids,
		creator: creator,
		now:     time.Now,
		logger:  logger,
	}
}

// Register creates a new account, persists it, and returns the result.
func (r *Registrar) Register(
	ctx context.Context,
	username, password string,
) (Account, error) {
	id, err := r.ids.NextID()
	if err != nil {
		return Account{}, fmt.Errorf(
			"could not generate account ID: %w",
			err,
		)
	}

	acc, err := Register(id, username, password, r.now())
	if err != nil {
		return Account{}, fmt.Errorf(
			"could not register account: %w",
			err,
		)
	}

	if err := r.creator.Create(ctx, &acc); err != nil {
		return Account{}, fmt.Errorf(
			"could not create account: %w",
			err,
		)
	}

	r.logger.Info(
		"account created",
		"account_id", acc.id,
		"username", acc.username.Value(),
		"display_name", acc.displayName,
		"role", acc.role,
		"status", acc.status,
		"created_at", acc.createdAt.Format(time.RFC3339),
	)

	return acc, nil
}
