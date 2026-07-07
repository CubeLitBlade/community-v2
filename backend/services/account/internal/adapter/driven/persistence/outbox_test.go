package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/adapter/driven/persistence"
)

func TestOutboxRepository_RecordAndScan(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := persistence.NewOutboxRepository(db)
	ctx := context.Background()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	entry := outbox.NewEntry(1, "/community-v2/account-service",
		"account.created.v1", "io.github.cubelitblade.account.created.v1", now,
		outbox.WithPayload(map[string]string{"key": "value"}),
	)

	if err := repo.Record(ctx, entry); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	entries, err := repo.Scan(ctx, db, 10)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Scan() returned %d entries, want 1", len(entries))
	}

	scanned := entries[0]
	if scanned.ID != 1 {
		t.Errorf("ID = %d, want 1", scanned.ID)
	}
	if scanned.Topic != "account.created.v1" {
		t.Errorf("Topic = %q, want account.created.v1", scanned.Topic)
	}
}

func TestOutboxRepository_Ack(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := persistence.NewOutboxRepository(db)
	ctx := context.Background()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	entry := outbox.NewEntry(2, "/community-v2/account-service",
		"account.created.v1", "io.github.cubelitblade.account.created.v1", now,
	)

	if err := repo.Record(ctx, entry); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	ackTime := now.Add(time.Minute)
	if err := repo.Ack(ctx, db, 2, ackTime); err != nil {
		t.Fatalf("Ack() error = %v", err)
	}

	entries, err := repo.Scan(ctx, db, 10)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf("Scan() returned %d entries after ack, want 0", len(entries))
	}
}

func TestOutboxRepository_ScanEmpty(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := persistence.NewOutboxRepository(db)
	ctx := context.Background()

	entries, err := repo.Scan(ctx, db, 10)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf("Scan() returned %d entries, want 0", len(entries))
	}
}
