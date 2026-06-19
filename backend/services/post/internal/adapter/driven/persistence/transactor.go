package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence/ent"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/port"
)

type EntTransactor struct {
	client *ent.Client
}

func NewEntTransactor(client *ent.Client) port.Transactor {
	return &EntTransactor{
		client: client,
	}
}

func (m *EntTransactor) InTx(ctx context.Context, fn func(context.Context) error) error {
	tx, err := m.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("starting tx: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			m.handlePanic(tx, v)
		}
	}()

	txCtx := ent.NewTxContext(ctx, tx)
	if err := fn(txCtx); err != nil {
		return m.rollbackOnError(tx, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing tx: %w", err)
	}

	return nil
}

func (*EntTransactor) handlePanic(tx *ent.Tx, v any) {
	if rerr := tx.Rollback(); rerr != nil {
		panic(fmt.Errorf("panic: %v, rollback also failed: %w", v, rerr))
	}

	panic(v)
}

func (*EntTransactor) rollbackOnError(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		return errors.Join(
			fmt.Errorf("executing tx: %w", err),
			fmt.Errorf("rolling back: %w", rerr),
		)
	}

	return fmt.Errorf("executing tx: %w", err)
}
