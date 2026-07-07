package account_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/account"
)

var errFinder = errors.New("finder error")

var testNow = time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)

type stubFinder struct {
	acc *account.Account
	err error
}

func (f *stubFinder) FindByUsername(_ context.Context, _ *gorm.DB, _ string) (*account.Account, error) {
	if f.err != nil {
		return nil, f.err
	}

	return f.acc, nil
}

func makeAccount(t *testing.T, password string) *account.Account {
	t.Helper()

	acc, err := account.NewAccount(1, "test_user", password, testNow)
	if err != nil {
		t.Fatalf("NewAccount() failed: %v", err)
	}

	return &acc
}

func stubNow() time.Time {
	return testNow
}

func newTestAuthenticator(finder account.ByUsernameFinder) *account.Authenticator {
	return account.NewAuthenticator(
		finder,
		nil,
		nil,
		account.WithAuthenticatorClock(stubNow),
	)
}

func TestAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()

	validAcc := makeAccount(t, validPwd)

	tests := []struct {
		name     string
		finder   account.ByUsernameFinder
		username string
		password string
		wantErr  error
	}{
		{
			name:     "Success",
			finder:   &stubFinder{acc: validAcc}, //nolint:exhaustruct // for testing only
			username: "test_user",
			password: validPwd,
			wantErr:  nil,
		},
		{
			name:     "AccountNotFound",
			finder:   &stubFinder{err: account.ErrAccountNotFound}, //nolint:exhaustruct // for testing only
			username: "missing",
			password: validPwd,
			wantErr:  account.ErrInvalidCredentials,
		},
		{
			name:     "WrongPassword",
			finder:   &stubFinder{acc: validAcc}, //nolint:exhaustruct // for testing only
			username: "test_user",
			password: "wrong-password!!!",
			wantErr:  account.ErrInvalidCredentials,
		},
		{
			name:     "NilAccount",
			finder:   &stubFinder{}, //nolint:exhaustruct // for testing only
			username: "ghost",
			password: validPwd,
			wantErr:  account.ErrInvalidCredentials,
		},
		{
			name:     "FinderError",
			finder:   &stubFinder{err: errFinder}, //nolint:exhaustruct // for testing only
			username: "any",
			password: validPwd,
			wantErr:  errFinder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			auth := newTestAuthenticator(tt.finder)

			_, err := auth.Authenticate(context.Background(), tt.username, tt.password)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Authenticate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
