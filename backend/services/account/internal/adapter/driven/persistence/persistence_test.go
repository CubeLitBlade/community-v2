package persistence_test

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE accounts (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			password_change_required INTEGER DEFAULT 0 NOT NULL,
			display_name TEXT NOT NULL,
			role TEXT DEFAULT 'member' NOT NULL,
			status TEXT DEFAULT 'active' NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			last_login_at DATETIME,
			last_login_ip TEXT
		)
	`).Error; err != nil {
		t.Fatalf("failed to create accounts table: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE outbox (
			id INTEGER PRIMARY KEY,
			aggregate_id INTEGER NOT NULL,
			aggregate_type TEXT NOT NULL,
			topic TEXT NOT NULL,
			event_type TEXT NOT NULL,
			payload TEXT,
			created_at DATETIME NOT NULL,
			published_at DATETIME,
			trace_id TEXT
		)
	`).Error; err != nil {
		t.Fatalf("failed to create outbox table: %v", err)
	}

	return db
}
