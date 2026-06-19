// Package authz provides a driven adapter for authorization operations.
// It implements the port.Authorizer interface by interacting with an external
// authorization service via gRPC to check user permissions.
package authz

import (
	"context"
	"fmt"

	authzv1 "github.com/cubelitblade/community-v2/backend/gen/go/authz/v1"
	"github.com/cubelitblade/community-v2/backend/services/authz/permission"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/port"
)

type Authorizer struct {
	client authzv1.AuthzServiceClient
}

func NewAuthorizer(client authzv1.AuthzServiceClient) port.Authorizer {
	return &Authorizer{client: client}
}

func (a *Authorizer) CanPublishPost(ctx context.Context, accountID int64) (bool, error) {
	resp, err := a.client.Check(ctx, &authzv1.CheckRequest{
		User:     fmt.Sprintf("user:%d", accountID),
		Relation: permission.RelationCanPublishPost,
		Object:   permission.ObjectSystemCommunity,
	})
	if err != nil {
		return false, fmt.Errorf("check permission: %w", err)
	}

	return resp.Allowed, nil
}
