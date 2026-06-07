package authz

import (
	"context"
	"fmt"

	sdk "github.com/openfga/go-sdk/client"
)

// TODO: currently the interface returns sdk.SdkClientGetStoreRequestInterface.
// should be refactored to pure domain interface, leaving OpenFGA just in bootstrap

// SDKStore defines the interface for the OpenFGA store client.
type SDKStore interface {
	GetStore(ctx context.Context) sdk.SdkClientGetStoreRequestInterface
}

// HealthChecker verifies the connectivity to the OpenFGA service.
type HealthChecker struct {
	client SDKStore
}

// NewHealthChecker creates a Checker backed by the given OpenFGA SDK client.
func NewHealthChecker(client SDKStore) *HealthChecker {
	return &HealthChecker{
		client: client,
	}
}

// Ping verifies connectivity to OpenFGA by reading the store.
func (c *HealthChecker) Ping(ctx context.Context) error {
	_, err := c.client.GetStore(ctx).Execute()
	if err != nil {
		return fmt.Errorf("openfga ping: %w", err)
	}

	return nil
}
