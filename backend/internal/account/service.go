package account

import (
	"fmt"
	"log/slog"
	"time"
)

type IDGenerator interface {
	NextID() (int64, error)
}

type Service struct {
	ids    IDGenerator
	now    func() time.Time
	logger *slog.Logger
}

func NewService(ids IDGenerator, logger *slog.Logger) *Service {
	if logger == nil {
		panic("account.NewService: logger is nil")
	}

	return &Service{
		ids:    ids,
		now:    time.Now,
		logger: logger,
	}
}

func (s *Service) CreateAccount(username string, password string) (Account, error) {
	id, err := s.ids.NextID()
	if err != nil {
		return Account{}, fmt.Errorf("generate account id: %w", err)
	}

	account, err := Register(ID(id), username, password, s.now())
	if err != nil {
		return Account{}, fmt.Errorf("create account: %w", err)
	}

	s.logger.Info(
		"account created",
		"account_id", account.id,
		"username", account.username.Value(),
		"display_name", account.displayName,
		"role", account.role,
		"status", account.status,
		"created_at", account.audit.createdAt.Format(time.RFC3339),
	)

	// TODO: save created account to database

	return account, nil
}
