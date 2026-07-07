package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/application"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain/port"
)

type stubIDGenerator struct {
	nextID int64
	err    error
}

func (s *stubIDGenerator) NextID() (int64, error) { return s.nextID, s.err }

type stubAccountWriter struct {
	created *account.Account
	err     error
}

func (s *stubAccountWriter) Create(_ context.Context, acc *account.Account) error {
	s.created = acc
	return s.err
}

func (s *stubAccountWriter) Update(_ context.Context, _ *account.Account) error {
	return nil
}

func TestRegistrar_Register(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	idGen := &stubIDGenerator{nextID: 42}
	writer := &stubAccountWriter{}

	r := application.NewRegistrar(idGen, writer)
	r.WithClock(func() time.Time { return fixedTime })

	acc, err := r.Register(context.Background(), "Alice", "this-is-a-valid-password")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if acc.ID != 42 {
		t.Errorf("ID() = %d, want 42", acc.ID)
	}

	if acc.Username != "Alice" {
		t.Errorf("Username() = %q, want %q", acc.Username, "Alice")
	}

	if writer.created == nil {
		t.Fatal("Create was not called")
	}

	if writer.created.ID != 42 {
		t.Errorf("created ID = %d, want 42", writer.created.ID)
	}
}

func TestRegistrar_Register_IDGeneratorError(t *testing.T) {
	t.Parallel()

	idGen := &stubIDGenerator{err: errors.New("snowflake down")}
	writer := &stubAccountWriter{}

	r := application.NewRegistrar(idGen, writer)

	_, err := r.Register(context.Background(), "Alice", "this-is-a-valid-password")
	if err == nil {
		t.Fatal("expected error from ID generator, got nil")
	}
}

func TestRegistrar_Register_InvalidPassword(t *testing.T) {
	t.Parallel()

	idGen := &stubIDGenerator{nextID: 1}
	writer := &stubAccountWriter{}

	r := application.NewRegistrar(idGen, writer)

	_, err := r.Register(context.Background(), "Alice", "short")
	if !errors.Is(err, account.ErrPasswordTooShort) {
		t.Errorf("error = %v, want %v", err, account.ErrPasswordTooShort)
	}
}

func TestRegistrar_Register_PersistError(t *testing.T) {
	t.Parallel()

	idGen := &stubIDGenerator{nextID: 1}
	writer := &stubAccountWriter{err: errors.New("db down")}

	r := application.NewRegistrar(idGen, writer)

	_, err := r.Register(context.Background(), "Alice", "this-is-a-valid-password")
	if err == nil {
		t.Fatal("expected persist error, got nil")
	}
}

func TestRegistrar_NewRegistrar_UsesRealtimeClock(t *testing.T) {
	r := application.NewRegistrar(nil, nil)
	if r == nil {
		t.Fatal("NewRegistrar returned nil")
	}
}

var _ port.IDGenerator = (*stubIDGenerator)(nil)
var _ port.AccountWriter = (*stubAccountWriter)(nil)
