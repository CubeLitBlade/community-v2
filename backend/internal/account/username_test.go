package account_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
)

func TestNewUsername(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    account.Username
		wantErr error
	}{
		{
			name:  "accepts minimum length username",
			value: strings.Repeat("a", account.MinUsernameLength),
			want:  account.Username(strings.Repeat("a", account.MinUsernameLength)),
		},
		{
			name:  "accepts maximum length username",
			value: strings.Repeat("a", account.MaxUsernameLength),
			want:  account.Username(strings.Repeat("a", account.MaxUsernameLength)),
		},
		{
			name:    "rejects username shorter than minimum length",
			value:   "a",
			wantErr: account.ErrUsernameLength,
		},
		{
			name:    "rejects username longer than maximum length",
			value:   strings.Repeat("a", account.MaxUsernameLength+1),
			wantErr: account.ErrUsernameLength,
		},
		{
			name:  "trims surrounding spaces",
			value: " alice ",
			want:  account.Username("alice"),
		},
		{
			name:    "rejects blank username",
			value:   "   ",
			wantErr: account.ErrUsernameBlank,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := account.NewUsername(tt.value)

			if tt.wantErr != nil {
				if gotErr == nil {
					t.Fatal("NewUsername() succeeded unexpectedly")
				}

				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("NewUsername() error = %v, want %v", gotErr, tt.wantErr)
				}

				return
			}

			if gotErr != nil {
				t.Fatalf("NewUsername() failed: %v", gotErr)
			}

			if got != tt.want {
				t.Errorf("NewUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUsernameFromStorage(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    account.Username
		wantErr error
	}{
		{
			name:  "accepts normal username",
			value: "bob",
			want:  account.Username("bob"),
		},
		{
			name:  "accepts archived username",
			value: "alice#archived_123",
			want:  account.Username("alice#archived_123"),
		},
		{
			name:  "accepts username longer than normal limit",
			value: strings.Repeat("a", account.MaxUsernameLength+1),
			want:  account.Username(strings.Repeat("a", account.MaxUsernameLength+1)),
		},
		{
			name:    "rejects blank stored username",
			value:   "   ",
			wantErr: account.ErrUsernameBlank,
		},
		{
			name:    "rejects username longer than storage limit",
			value:   strings.Repeat("a", account.MaxStoredUsernameLength+1),
			wantErr: account.ErrUsernameLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := account.UsernameFromStorage(tt.value)

			if tt.wantErr != nil {
				if gotErr == nil {
					t.Fatal("UsernameFromStorage() succeeded unexpectedly")
				}

				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("UsernameFromStorage() error = %v, want %v", gotErr, tt.wantErr)
				}

				return
			}

			if gotErr != nil {
				t.Fatalf("UsernameFromStorage() failed: %v", gotErr)
			}

			if got != tt.want {
				t.Errorf("UsernameFromStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArchiveUsername(t *testing.T) {
	tests := []struct {
		name     string
		username account.Username
		id       account.ID
		want     account.Username
		wantErr  error
	}{
		{
			name:     "generates archived username from normal username",
			username: account.Username("alice"),
			id:       account.ID(9223372036854775807),
			want:     account.Username("alice#archived_9223372036854775807"),
		},
		{
			name:     "keeps archived username within storage limit",
			username: account.Username(strings.Repeat("a", account.MaxUsernameLength)),
			id:       account.ID(9223372036854775807),
			want:     account.Username(strings.Repeat("a", account.MaxUsernameLength) + "#archived_9223372036854775807"),
		},
		{
			name:     "rejects archived username longer than storage limit",
			username: account.Username(strings.Repeat("a", account.MaxStoredUsernameLength)),
			id:       account.ID(1),
			wantErr:  account.ErrUsernameLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := account.ArchiveUsername(tt.username, tt.id)

			if tt.wantErr != nil {
				if gotErr == nil {
					t.Fatal("ArchiveUsername() succeeded unexpectedly")
				}

				if !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("ArchiveUsername() error = %v, want %v", gotErr, tt.wantErr)
				}

				return
			}

			if gotErr != nil {
				t.Fatalf("ArchiveUsername() failed: %v", gotErr)
			}

			if got != tt.want {
				t.Errorf("ArchiveUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}
