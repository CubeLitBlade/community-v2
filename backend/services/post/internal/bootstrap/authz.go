package bootstrap

import (
	"go.uber.org/fx"
	"google.golang.org/grpc"

	authzv1 "github.com/cubelitblade/community-v2/backend/gen/go/authz/v1"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/authz"
)

func NewAuthzServiceClient(conn *grpc.ClientConn) authzv1.AuthzServiceClient {
	return authzv1.NewAuthzServiceClient(conn)
}

func AuthzModule() fx.Option {
	return fx.Options(
		fx.Provide(authz.NewAuthorizer),
		fx.Provide(NewAuthzServiceClient),
	)
}
