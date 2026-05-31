package account

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
	"gorm.io/gorm"
)

// RegistrarOption defines a functional option for configuring a Registrar.
type RegistrarOption func(*Registrar)

// Creator persists a new account.
type Creator interface {
	Create(ctx context.Context, db *gorm.DB, acc *Account) error
}

// OutboxSaver persists an outbox event for asynchronous processing.
type OutboxSaver interface {
	Save(ctx context.Context, db *gorm.DB, event *outbox.Entry) error
}

// Registrar orchestrates the account registration process.
type Registrar struct {
	ids         idgen.Generator
	creator     Creator
	outboxSaver OutboxSaver
	db          *gorm.DB
	clock       func() time.Time
	logger      *slog.Logger
}

// NewRegistrar creates and returns a new Registrar.
func NewRegistrar(ids idgen.Generator, creator Creator, outboxSaver OutboxSaver,
	db *gorm.DB, logger *slog.Logger, opts ...RegistrarOption,
) *Registrar {
	if logger == nil {
		logger = slog.Default()
		logger.Warn("No logger provided while creating registrar, using default logger.")
	}

	registrar := &Registrar{
		ids:         ids,
		creator:     creator,
		outboxSaver: outboxSaver,
		db:          db,
		clock:       time.Now,
		logger:      logger,
	}

	for _, opt := range opts {
		opt(registrar)
	}

	return registrar
}

// Register creates a new account, persists it, and returns the result.
func (r *Registrar) Register(ctx context.Context, username, password string) (Account, error) {
	now := r.clock()

	id, err := r.ids.NextID()
	if err != nil {
		r.logger.Error("Failed to generate account id.", "error", err)

		return Account{}, fmt.Errorf("generate ID: %w", err)
	}

	acc, err := NewAccount(id, username, password, now)
	if err != nil {
		r.logger.Error("Failed to create account.", "error", err)

		return Account{}, fmt.Errorf("create account: %w", err)
	}

	id, err = r.ids.NextID()
	if err != nil {
		r.logger.Error("Failed to generate account id.", "error", err)

		return Account{}, fmt.Errorf("generate ID: %w", err)
	}

	event := outbox.NewEntry(
		id, "/community-v2/account-service", "account.created", now,
		outbox.WithPayload(map[string]any{
			"id":           strconv.FormatInt(acc.ID(), 10),
			"username":     acc.Username(),
			"display_name": acc.DisplayName(),
			"role":         acc.Role(),
			"status":       acc.Status(),
			"created_at":   acc.createdAt.Format(time.RFC3339),
		}),
	)

	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := r.creator.Create(ctx, tx, &acc); err != nil {
			r.logger.Error("Failed to persist account.", "error", err)

			return fmt.Errorf("persist account: %w", err)
		}

		if err := r.outboxSaver.Save(ctx, tx, event); err != nil {
			r.logger.Error("Failed to persist event.", "error", err)

			return fmt.Errorf("persist event: %w", err)
		}

		return nil
	})
	if err != nil {
		r.logger.Error("Transaction failed.", "error", err)

		return Account{}, fmt.Errorf("transaction: %w", err)
	}

	return acc, nil
}

// WithRegistrarClock sets the clock function for the Registrar.
func WithRegistrarClock(clock func() time.Time) RegistrarOption {
	return func(r *Registrar) {
		r.clock = clock
	}
}
