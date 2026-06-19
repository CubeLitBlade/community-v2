package openfga

import (
	"context"
	"fmt"
	"log/slog"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/config"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

type Writer struct {
	cfg    config.OpenFGAConfig
	client openfgav1.OpenFGAServiceClient
	logger *slog.Logger
}

func NewWriter(cfg config.OpenFGAConfig, client openfgav1.OpenFGAServiceClient, logger *slog.Logger) port.TupleWriter {
	return &Writer{
		cfg:    cfg,
		client: client,
		logger: logger,
	}
}

func (w *Writer) Write(ctx context.Context, tuples []domain.Tuple) error {
	tupleKeys := make([]*openfgav1.TupleKey, 0, len(tuples))
	for _, t := range tuples {
		tupleKeys = append(tupleKeys, &openfgav1.TupleKey{
			User:     t.User,
			Relation: t.Relation,
			Object:   t.Object,
		})
	}

	_, err := w.client.Write(ctx, &openfgav1.WriteRequest{
		StoreId: w.cfg.StoreID,
		Writes: &openfgav1.WriteRequestWrites{
			TupleKeys: tupleKeys,
		},
		AuthorizationModelId: w.cfg.ModelID,
	})
	if err != nil {
		return fmt.Errorf("write tuple keys: %w", err)
	}

	return nil
}
