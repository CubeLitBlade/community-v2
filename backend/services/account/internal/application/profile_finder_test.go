package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/application"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
)

func TestProfileFinder_Find_Success(t *testing.T) {
	t.Parallel()

	acc, err := account.NewAccount(42, "Alice", "this-is-a-valid-password", time.Now())
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	reader := &stubAccountReader{byID: &acc}
	f := application.NewProfileFinder(reader)

	profile, err := f.Find(context.Background(), 42)
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}

	if profile.ID != 42 {
		t.Errorf("ID = %d, want 42", profile.ID)
	}
	if profile.Username != "Alice" {
		t.Errorf("Username = %q, want %q", profile.Username, "Alice")
	}
	if profile.DisplayName != "Alice" {
		t.Errorf("DisplayName = %q, want %q", profile.DisplayName, "Alice")
	}
}

func TestProfileFinder_Find_NotFound(t *testing.T) {
	t.Parallel()

	reader := &stubAccountReader{byIDErr: account.ErrAccountNotFound}
	f := application.NewProfileFinder(reader)

	_, err := f.Find(context.Background(), 999)
	if !errors.Is(err, account.ErrAccountNotFound) {
		t.Errorf("error = %v, want %v", err, account.ErrAccountNotFound)
	}
}
