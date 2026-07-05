package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	authzv1 "github.com/cubelitblade/community-v2/backend/gen/go/authz/v1"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/application"
)

type AuthzServiceServer struct {
	authzv1.UnimplementedAuthzServiceServer

	check application.AuthChecker
}

func NewAuthzServiceServer(check application.AuthChecker) *AuthzServiceServer {
	return &AuthzServiceServer{
		UnimplementedAuthzServiceServer: authzv1.UnimplementedAuthzServiceServer{},
		check:                           check,
	}
}

func (s *AuthzServiceServer) Check(ctx context.Context, req *authzv1.CheckRequest) (*authzv1.CheckResponse, error) {
	allowed, err := s.check.Authorize(ctx, req.User, req.Relation, req.Object)
	if err != nil {
		return nil, fmt.Errorf("check permission: %w", err)
	}

	return &authzv1.CheckResponse{
		Allowed: allowed,
	}, nil
}

func (s *AuthzServiceServer) RegisterGRPC(srv *grpc.Server) {
	authzv1.RegisterAuthzServiceServer(srv, s)
}
