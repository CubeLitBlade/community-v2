package account

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
	v1 "github.com/cubelitblade/community-v2/backend/services/account/api/events/v1"
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
	logger = logger.With(
		slog.String("service", "account"),
		slog.String("component", "account/registrar"),
	)

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

// Register creates a new account, persists it, and returns the ID.
func (r *Registrar) Register(ctx context.Context, username, password string) (int64, error) {
	now := r.clock()

	id, err := r.ids.NextID()
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to generate account id",
			slog.Any("error", err),
		)

		return 0, fmt.Errorf("generate ID: %w", err)
	}

	acc, err := NewAccount(id, username, password, now)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to create account",
			slog.Any("error", err),
		)

		return 0, fmt.Errorf("create account: %w", err)
	}

	id, err = r.ids.NextID()
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to generate event id",
			slog.Any("error", err),
		)

		return 0, fmt.Errorf("generate ID: %w", err)
	}

	entry := newAccountCreatedEntry(id, acc, now)

	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := r.creator.Create(ctx, tx, &acc); err != nil {
			r.logger.ErrorContext(ctx, "failed to persist account",
				slog.Any("error", err),
			)

			return fmt.Errorf("persist account: %w", err)
		}

		if err := r.outboxSaver.Save(ctx, tx, entry); err != nil {
			r.logger.ErrorContext(ctx, "failed to persist event",
				slog.Any("error", err))

			return fmt.Errorf("persist event: %w", err)
		}

		return nil
	})
	if err != nil {
		r.logger.ErrorContext(ctx, "transaction failed",
			slog.Any("error", err),
		)

		return 0, fmt.Errorf("transaction: %w", err)
	}

	return acc.ID(), nil
}

// WithRegistrarClock sets the clock function for the Registrar.
func WithRegistrarClock(clock func() time.Time) RegistrarOption {
	return func(r *Registrar) {
		r.clock = clock
	}
}

func newAccountCreatedEntry(eventID int64, acc Account, now time.Time) *outbox.Entry {
	return outbox.NewEntry(eventID, v1.AggregateTypeAccountService,
		v1.TopicAccountCreated, v1.EventTypeAccountCreated, now, outbox.WithPayload(
			v1.AccountCreatedEventPayload{
				AccountID: strconv.FormatInt(acc.ID(), 10),
				Username:  acc.Username(),
				Role:      string(acc.role),
				CreatedAt: acc.createdAt.Format(time.RFC3339),
			},
		),
	)
}
