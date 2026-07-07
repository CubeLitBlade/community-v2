package grpc_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	accountv1 "github.com/cubelitblade/community-v2/backend/gen/go/account/v1"
	accountgrpc "github.com/cubelitblade/community-v2/backend/services/account/internal/adapter/driving/grpc"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/application"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
)

const bufSize = 1024 * 1024

type testRegistrar struct {
	acc *account.Account
	err error
}

func (r *testRegistrar) Register(_ context.Context, _, _ string) (*account.Account, error) {
	return r.acc, r.err
}

type testAuthenticator struct {
	acc *account.Account
	err error
}

func (a *testAuthenticator) Authenticate(_ context.Context, _, _ string) (*account.Account, error) {
	return a.acc, a.err
}

type testProfileFinder struct {
	profile *application.Profile
	err     error
}

func (f *testProfileFinder) Find(_ context.Context, _ int64) (*application.Profile, error) {
	return f.profile, f.err
}

func setupServer(t *testing.T, registrar application.RegistrarUseCase, authenticator application.AuthenticatorUseCase, finder application.ProfileFinderUseCase) (accountv1.AccountServiceClient, func()) {
	t.Helper()

	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()

	handler := accountgrpc.NewAccountServiceServer(registrar, authenticator, finder)
	handler.RegisterGRPC(srv)

	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Logf("grpc server stopped: %v", err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient("passthrough://bufnet", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}

	client := accountv1.NewAccountServiceClient(conn)

	cleanup := func() {
		conn.Close()
		srv.GracefulStop()
	}

	return client, cleanup
}

func mustNewAccount(t *testing.T) *account.Account {
	t.Helper()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	acc, err := account.NewAccount(42, "Alice", "this-is-a-valid-password", now)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	return &acc
}

func TestRegister(t *testing.T) {
	acc := mustNewAccount(t)
	registrar := &testRegistrar{acc: acc}

	client, cleanup := setupServer(t, registrar, &testAuthenticator{}, &testProfileFinder{})
	defer cleanup()

	resp, err := client.Register(context.Background(), &accountv1.RegisterRequest{
		Username: "Alice",
		Password: "this-is-a-valid-password",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if resp.AccountId != int64(acc.ID) {
		t.Errorf("AccountId = %d, want %d", resp.AccountId, int64(acc.ID))
	}
}

func TestRegister_Error(t *testing.T) {
	registrar := &testRegistrar{err: errors.New("registration failed")}

	client, cleanup := setupServer(t, registrar, &testAuthenticator{}, &testProfileFinder{})
	defer cleanup()

	_, err := client.Register(context.Background(), &accountv1.RegisterRequest{
		Username: "Alice",
		Password: "this-is-a-valid-password",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAuthenticate(t *testing.T) {
	acc := mustNewAccount(t)
	auth := &testAuthenticator{acc: acc}

	client, cleanup := setupServer(t, &testRegistrar{}, auth, &testProfileFinder{})
	defer cleanup()

	resp, err := client.Authenticate(context.Background(), &accountv1.AuthenticateRequest{
		Username: "Alice",
		Password: "this-is-a-valid-password",
	})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if resp.AccountId != int64(acc.ID) {
		t.Errorf("AccountId = %d, want %d", resp.AccountId, int64(acc.ID))
	}
	if resp.Role != "member" {
		t.Errorf("Role = %q, want member", resp.Role)
	}
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	auth := &testAuthenticator{err: account.ErrInvalidCredentials}

	client, cleanup := setupServer(t, &testRegistrar{}, auth, &testProfileFinder{})
	defer cleanup()

	_, err := client.Authenticate(context.Background(), &accountv1.AuthenticateRequest{
		Username: "Alice",
		Password: "wrong",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetProfile(t *testing.T) {
	finder := &testProfileFinder{
		profile: &application.Profile{
			ID:          42,
			Username:    "Alice",
			DisplayName: "Alice",
		},
	}

	client, cleanup := setupServer(t, &testRegistrar{}, &testAuthenticator{}, finder)
	defer cleanup()

	resp, err := client.GetProfile(context.Background(), &accountv1.GetProfileRequest{
		AccountId: 42,
	})
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if resp.Id != 42 {
		t.Errorf("Id = %d, want 42", resp.Id)
	}
	if resp.Username != "Alice" {
		t.Errorf("Username = %q, want Alice", resp.Username)
	}
	if resp.DisplayName != "Alice" {
		t.Errorf("DisplayName = %q, want Alice", resp.DisplayName)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	finder := &testProfileFinder{err: account.ErrAccountNotFound}

	client, cleanup := setupServer(t, &testRegistrar{}, &testAuthenticator{}, finder)
	defer cleanup()

	_, err := client.GetProfile(context.Background(), &accountv1.GetProfileRequest{
		AccountId: 999,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
