package authz

import (
	"context"
	"fmt"
	"log/slog"

	sdk "github.com/openfga/go-sdk/client"
)

// TODO: currently the interface returns sdk.SdkClientCheckRequestInterface.
// should be refactored to pure domain interface, such as PermissionChecker(ctx, user, relation, object) (bool, error)
// leaving OpenFGA just in bootstrap

// SDKChecker defines the interface for the OpenFGA check client.
type SDKChecker interface {
	Check(ctx context.Context) sdk.SdkClientCheckRequestInterface
}

// Checker performs authorization checks against OpenFGA.
type Checker struct {
	client SDKChecker
	logger *slog.Logger
}

// NewChecker creates a Checker backed by the given OpenFGA SDK client.
func NewChecker(client SDKChecker, logger *slog.Logger) *Checker {
	return &Checker{
		client: client,
		logger: logger,
	}
}

// Check returns whether the given user has the given relation to the given object.
func (c *Checker) Check(ctx context.Context, user, relation, object string) (bool, error) {
	body := sdk.ClientCheckRequest{
		User:     user,
		Relation: relation,
		Object:   object,
	}

	resp, err := c.client.Check(ctx).Body(body).Execute()
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to request openfga",
			slog.String("user", user),
			slog.String("relation", relation),
			slog.String("object", object),
			slog.Any("error", err),
		)

		return false, fmt.Errorf("openfga check: %w", err)
	}

	if resp.Allowed == nil {
		return false, nil
	}

	return *resp.Allowed, nil
}
