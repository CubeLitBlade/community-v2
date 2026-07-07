package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	accountv1 "github.com/cubelitblade/community-v2/backend/pkg/events/account/v1"
	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain/port"
)

type RegistrarUseCase interface {
	Register(ctx context.Context, username, password string) (*account.Account, error)
}

type Registrar struct {
	ids           port.IDGenerator
	creator       port.AccountWriter
	eventRecorder port.EventRecorder
	tx            port.TxRunner
	clock         func() time.Time
}

func NewRegistrar(
	ids port.IDGenerator,
	creator port.AccountWriter,
	eventRecorder port.EventRecorder,
	tx port.TxRunner,
) *Registrar {
	return &Registrar{
		ids:           ids,
		creator:       creator,
		eventRecorder: eventRecorder,
		tx:            tx,
		clock:         time.Now,
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

	entry := buildOutboxEntry(id, acc, now)

	if err := r.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := r.creator.Create(ctx, &acc); err != nil {
			return fmt.Errorf("persist account: %w", err)
		}

		if err := r.eventRecorder.Record(ctx, entry); err != nil {
			return fmt.Errorf("persist event: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("transaction: %w", err)
	}

	return &acc, nil
}

func (r *Registrar) WithClock(clock func() time.Time) *Registrar {
	r.clock = clock
	return r
}

func buildOutboxEntry(id int64, acc account.Account, now time.Time) *outbox.Entry {
	return outbox.NewEntry(id, accountv1.AggregateType,
		accountv1.TopicAccountCreated, accountv1.EventTypeAccountCreated, now, outbox.WithPayload(
			accountv1.AccountCreated{
				AccountId: strconv.FormatInt(int64(acc.ID), 10),
				Username:  acc.Username.Value(),
				Role:      acc.Role.String(),
				CreatedAt: acc.CreatedAt.Format(time.RFC3339),
			},
		),
	)
}
