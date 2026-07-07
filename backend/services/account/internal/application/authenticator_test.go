package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/application"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
)

type stubAccountReader struct {
	byID       *account.Account
	byIDErr    error
	byUsername *account.Account
	byUserErr  error
}

func (s *stubAccountReader) FindByID(_ context.Context, _ int64) (*account.Account, error) {
	return s.byID, s.byIDErr
}

func (s *stubAccountReader) FindByUsername(_ context.Context, _ account.Username) (*account.Account, error) {
	return s.byUsername, s.byUserErr
}

func mustCreateAccount(t *testing.T, username, password string) *account.Account {
	t.Helper()

	acc, err := account.NewAccount(1, username, password, time.Now())
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	return &acc
}

func TestAuthenticator_Authenticate_Success(t *testing.T) {
	t.Parallel()

	acc := mustCreateAccount(t, "Alice", "this-is-a-valid-password")
	reader := &stubAccountReader{byUsername: acc}
	a := application.NewAuthenticator(reader)

	result, err := a.Authenticate(context.Background(), "Alice", "this-is-a-valid-password")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if result.ID != 1 {
		t.Errorf("ID() = %d, want 1", result.ID)
	}
}

func TestAuthenticator_Authenticate_WrongPassword(t *testing.T) {
	t.Parallel()

	acc := mustCreateAccount(t, "Alice", "this-is-a-valid-password")
	reader := &stubAccountReader{byUsername: acc}
	a := application.NewAuthenticator(reader)

	_, err := a.Authenticate(context.Background(), "Alice", "wrong-password")
	if !errors.Is(err, account.ErrInvalidCredentials) {
		t.Errorf("error = %v, want %v", err, account.ErrInvalidCredentials)
	}
}

func TestAuthenticator_Authenticate_AccountNotFound(t *testing.T) {
	t.Parallel()

	reader := &stubAccountReader{byUserErr: account.ErrAccountNotFound}
	a := application.NewAuthenticator(reader)

	_, err := a.Authenticate(context.Background(), "Nobody", "password-here")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAuthenticator_Authenticate_InvalidUsername(t *testing.T) {
	t.Parallel()

	reader := &stubAccountReader{}
	a := application.NewAuthenticator(reader)

	_, err := a.Authenticate(context.Background(), "", "password-here")
	if !errors.Is(err, account.ErrInvalidCredentials) {
		t.Errorf("error = %v, want %v", err, account.ErrInvalidCredentials)
	}
}
