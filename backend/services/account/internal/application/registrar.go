package application

import (
	"context"
	"fmt"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain/port"
)

type RegistrarUseCase interface {
	Register(ctx context.Context, username, password string) (*account.Account, error)
}

type Registrar struct {
	ids     port.IDGenerator
	creator port.AccountWriter
	clock   func() time.Time
}

func NewRegistrar(ids port.IDGenerator, creator port.AccountWriter) *Registrar {
	return &Registrar{
		ids:     ids,
		creator: creator,
		clock:   time.Now,
	}
}

func (r *Registrar) Register(ctx context.Context, username, password string) (*account.Account, error) {
	now := r.clock()

	id, err := r.ids.NextID()
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	acc, err := account.NewAccount(id, username, password, now)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	if err := r.creator.Create(ctx, &acc); err != nil {
		return nil, fmt.Errorf("persist account: %w", err)
	}

	return &acc, nil
}

func (r *Registrar) WithClock(clock func() time.Time) *Registrar {
	r.clock = clock
	return r
}
