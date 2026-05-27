package account_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/account"
)

var errFinder = errors.New("finder error")

var testNow = time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)

type stubFinder struct {
	acc *account.Account
	err error
}

func (f *stubFinder) FindByUsername(
	_ context.Context, _ string,
) (*account.Account, error) {
	if f.err != nil {
		return nil, f.err
	}

	return f.acc, nil
}

func makeAccount(t *testing.T, password string) *account.Account {
	t.Helper()

	acc, err := account.Register(1, "testuser", password, testNow)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	return &acc
}

func stubNow() time.Time {
	return testNow
}

func TestAuthenticator_Success(t *testing.T) {
	t.Parallel()

	pwd := validPwd
	acc := makeAccount(t, pwd)
	finder := &stubFinder{acc: acc, err: nil}
	auth := account.NewAuthenticator(finder, stubNow, nil)

	got, err := auth.Authenticate(
		context.Background(), "testuser", pwd,
	)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if got.ID() != acc.ID() {
		t.Errorf("ID() = %v, want %v", got.ID(), acc.ID())
	}
}

func TestAuthenticator_AccountNotFound(t *testing.T) {
	t.Parallel()

	finder := &stubFinder{acc: nil, err: account.ErrAccountNotFound}
	auth := account.NewAuthenticator(finder, stubNow, nil)

	_, err := auth.Authenticate(
		context.Background(), "missing", validPwd,
	)
	if !errors.Is(err, account.ErrInvalidCredentials) {
		t.Errorf("error = %v, want %v",
			err, account.ErrInvalidCredentials)
	}
}

func TestAuthenticator_WrongPassword(t *testing.T) {
	t.Parallel()

	acc := makeAccount(t, validPwd)
	finder := &stubFinder{acc: acc, err: nil}
	auth := account.NewAuthenticator(finder, stubNow, nil)

	_, err := auth.Authenticate(
		context.Background(), "testuser", "wrong-password!!!",
	)
	if !errors.Is(err, account.ErrInvalidCredentials) {
		t.Errorf("error = %v, want %v",
			err, account.ErrInvalidCredentials)
	}
}

func TestAuthenticator_NilAccount(t *testing.T) {
	t.Parallel()

	finder := &stubFinder{acc: nil, err: nil}
	auth := account.NewAuthenticator(finder, stubNow, nil)

	_, err := auth.Authenticate(
		context.Background(), "ghost", validPwd,
	)
	if !errors.Is(err, account.ErrInvalidCredentials) {
		t.Errorf("error = %v, want %v",
			err, account.ErrInvalidCredentials)
	}
}

func TestAuthenticator_FinderError(t *testing.T) {
	t.Parallel()

	finder := &stubFinder{acc: nil, err: errFinder}
	auth := account.NewAuthenticator(finder, stubNow, nil)

	_, err := auth.Authenticate(
		context.Background(), "any", validPwd,
	)
	if !errors.Is(err, errFinder) {
		t.Errorf("error = %v, want %v", err, errFinder)
	}
}
