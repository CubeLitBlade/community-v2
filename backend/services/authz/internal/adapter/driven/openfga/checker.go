package openfga

import (
	"context"
	"fmt"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/config"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

// Checker is an adapter that evaluates permissions using an OpenFGA backend.
type Checker struct {
	cfg    config.OpenFGAConfig
	client openfgav1.OpenFGAServiceClient
}

// NewChecker creates a new OpenFGA authority checker.
func NewChecker(cfg config.OpenFGAConfig, client openfgav1.OpenFGAServiceClient) port.AuthorityChecker {
	return &Checker{
		cfg:    cfg,
		client: client,
	}
}

// Check queries OpenFGA to determine if a user is granted a relation on an object.
func (c *Checker) Check(ctx context.Context, user, relation, object string) (bool, error) {
	resp, err := c.client.Check(ctx, &openfgav1.CheckRequest{
		StoreId: c.cfg.StoreID,
		TupleKey: &openfgav1.CheckRequestTupleKey{
			User:     user,
			Relation: relation,
			Object:   object,
		},
		AuthorizationModelId: c.cfg.ModelID,
		Trace:                true,
	})
	if err != nil {
		return false, fmt.Errorf("request openfga: %w", err)
	}

	return resp.Allowed, nil
}
