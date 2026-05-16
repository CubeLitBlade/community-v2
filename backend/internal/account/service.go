package account

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// ErrUsernameAlreadyExists is returned when attempting to create an account
// with a username that is already taken.
var ErrUsernameAlreadyExists = errors.New("username already exists")

// IDGenerator generates unique identifiers for new accounts.
type IDGenerator interface {
	NextID() (int64, error)
}

// Writer defines the interface for persisting account aggregates.
type Writer interface {
	Create(ctx context.Context, acc *Account) error
}

// Reader defines the interface for querying account aggregates.
type Reader interface {
	LookupByUsername(ctx context.Context, username string) (Account, error)
	ExistsUsername(ctx context.Context, username string) (bool, error)
}

// Service provides business logic for account operations.
type Service struct {
	ids    IDGenerator
	writer Writer
	reader Reader
	now    func() time.Time
	logger *slog.Logger
}

// NewService creates a new Service with the given dependencies.
// It panics if logger is nil.
func NewService(
	ids IDGenerator,
	writer Writer,
	reader Reader,
	logger *slog.Logger,
) *Service {
	if logger == nil {
		panic("account.NewService: logger is nil")
	}

	return &Service{
		ids:    ids,
		writer: writer,
		reader: reader,
		now:    time.Now,
		logger: logger,
	}
}

// CreateAccount registers a new account, persists it, and logs the creation.
func (s *Service) CreateAccount(
	ctx context.Context,
	username, password string,
) (Account, error) {
	existsUsername, err := s.reader.ExistsUsername(ctx, username)
	if err != nil {
		return Account{}, fmt.Errorf("checking account exists: %w", err)
	}
	if existsUsername {
		return Account{}, ErrUsernameAlreadyExists
	}

	id, err := s.ids.NextID()
	if err != nil {
		return Account{}, fmt.Errorf("generate account id: %w", err)
	}

	acc, err := Register(id, username, password, s.now())
	if err != nil {
		return Account{}, fmt.Errorf("create account: %w", err)
	}

	if err := s.writer.Create(ctx, &acc); err != nil {
		return Account{}, fmt.Errorf("create account: %w", err)
	}

	s.logger.Info(
		"account created",
		"account_id", acc.id,
		"username", acc.username.Value(),
		"display_name", acc.displayName,
		"role", acc.role,
		"status", acc.status,
		"created_at", acc.audit.createdAt.Format(time.RFC3339),
	)

	return acc, nil
}
