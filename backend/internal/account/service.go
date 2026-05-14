package account

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type IDGenerator interface {
	NextID() (int64, error)
}

type repository interface {
	Create(ctx context.Context, acc *Account) error
}

type Service struct {
	ids    IDGenerator
	repo   repository
	now    func() time.Time
	logger *slog.Logger
}

func NewService(ids IDGenerator, repo repository, logger *slog.Logger) *Service {
	if logger == nil {
		panic("account.NewService: logger is nil")
	}

	return &Service{
		ids:    ids,
		repo:   repo,
		now:    time.Now,
		logger: logger,
	}
}

func (s *Service) CreateAccount(ctx context.Context, username, password string) (Account, error) {
	id, err := s.ids.NextID()
	if err != nil {
		return Account{}, fmt.Errorf("generate account id: %w", err)
	}

	acc, err := Register(ID(id), username, password, s.now())
	if err != nil {
		return Account{}, fmt.Errorf("create account: %w", err)
	}

	if err := s.repo.Create(ctx, &acc); err != nil {
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
