package persistence

import (
	"context"

	"gorm.io/gorm"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain/port"
)

type txKeyType struct{}

var txKey txKeyType

func NewTxRunner(db *gorm.DB) port.TxRunner {
	return &txRunner{db: db}
}

type txRunner struct {
	db *gorm.DB
}

func (t *txRunner) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}

func txFromContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}

	return db.WithContext(ctx)
}
