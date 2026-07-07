package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	accountv1 "github.com/cubelitblade/community-v2/backend/gen/go/account/v1"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/application"
)

type AccountServiceServer struct {
	accountv1.UnimplementedAccountServiceServer

	registrar     application.RegistrarUseCase
	authenticator application.AuthenticatorUseCase
	profileFinder application.ProfileFinderUseCase
}

func NewAccountServiceServer(
	registrar application.RegistrarUseCase,
	authenticator application.AuthenticatorUseCase,
	profileFinder application.ProfileFinderUseCase,
) *AccountServiceServer {
	return &AccountServiceServer{
		UnimplementedAccountServiceServer: accountv1.UnimplementedAccountServiceServer{},
		registrar:                         registrar,
		authenticator:                     authenticator,
		profileFinder:                     profileFinder,
	}
}

func (srv *AccountServiceServer) Register(ctx context.Context, req *accountv1.RegisterRequest) (*accountv1.RegisterResponse, error) {
	acc, err := srv.registrar.Register(ctx, req.Username, req.Password)
	if err != nil {
		return nil, fmt.Errorf("register account: %w", err)
	}

	return &accountv1.RegisterResponse{
		AccountId: int64(acc.ID),
	}, nil
}

func (srv *AccountServiceServer) Authenticate(ctx context.Context, req *accountv1.AuthenticateRequest) (*accountv1.AuthenticateResponse, error) {
	acc, err := srv.authenticator.Authenticate(ctx, req.Username, req.Password)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	return &accountv1.AuthenticateResponse{
		AccountId: int64(acc.ID),
		Role:      acc.Role.String(),
	}, nil
}

func (srv *AccountServiceServer) GetProfile(ctx context.Context, req *accountv1.GetProfileRequest) (*accountv1.GetProfileResponse, error) {
	profile, err := srv.profileFinder.Find(ctx, req.AccountId)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}

	return &accountv1.GetProfileResponse{
		Id:          profile.ID,
		Username:    profile.Username,
		DisplayName: profile.DisplayName,
	}, nil
}

func (srv *AccountServiceServer) RegisterGRPC(s *grpc.Server) {
	accountv1.RegisterAccountServiceServer(s, srv)
}
