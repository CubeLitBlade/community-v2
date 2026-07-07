package persistence

import (
	"context"

	"gorm.io/gorm"
)

type txKeyType struct{}

var txKey txKeyType

func NewTxRunner(db *gorm.DB) *TxRunner {
	return &TxRunner{db: db}
}

type TxRunner struct {
	db *gorm.DB
}

func (t *TxRunner) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
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
