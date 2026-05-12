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
		wantErr bool
	}{
		{
			name:    "accepts minimum length username",
			value:   strings.Repeat("a", account.MinUsernameLength),
			want:    account.Username(strings.Repeat("a", account.MinUsernameLength)),
			wantErr: false,
		},
		{
			name:    "accepts maximum length username",
			value:   strings.Repeat("a", account.MaxUsernameLength),
			want:    account.Username(strings.Repeat("a", account.MaxUsernameLength)),
			wantErr: false,
		},
		{
			name:    "rejects username shorter than minimum length",
			value:   "a",
			wantErr: true,
		},
		{
			name:    "rejects username longer than maximum length",
			value:   strings.Repeat("a", account.MaxUsernameLength+1),
			wantErr: true,
		},
		{
			name:    "trims surrounding spaces",
			value:   " alice ",
			want:    account.Username("alice"),
			wantErr: false,
		},
		{
			name:    "rejects blank username",
			value:   "   ",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := account.NewUsername(tt.value)
			if tt.wantErr {
				if gotErr == nil {
					t.Fatal("NewUsername() succeeded unexpectedly")
				}

				if !errors.Is(gotErr, account.ErrInvalidUsername) {
					t.Errorf("NewUsername() error = %v, want ErrInvalidUsername", gotErr)
				}

				return
			}
			if gotErr != nil {
				t.Errorf("NewUsername() failed: %v", gotErr)

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
		wantErr bool
	}{
		{
			name:    "accepts normal username",
			value:   "bob",
			want:    account.Username("bob"),
			wantErr: false,
		},
		{
			name:    "accepts archived username",
			value:   "alice#archived_123",
			want:    account.Username("alice#archived_123"),
			wantErr: false,
		},
		{
			name:    "accepts username longer than normal limit",
			value:   strings.Repeat("a", account.MaxUsernameLength+1),
			want:    account.Username(strings.Repeat("a", account.MaxUsernameLength+1)),
			wantErr: false,
		},
		{
			name:    "rejects blank stored username",
			value:   "   ",
			wantErr: true,
		},
		{
			name:    "rejects username longer than storage limit",
			value:   strings.Repeat("a", account.MaxStoredUsernameLength+1),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := account.UsernameFromStorage(tt.value)

			if tt.wantErr {
				if gotErr == nil {
					t.Fatal("UsernameFromStorage() succeeded unexpectedly")
				}

				if !errors.Is(gotErr, account.ErrInvalidUsername) {
					t.Errorf("UsernameFromStorage() error = %v, want ErrInvalidUsername", gotErr)
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
		id       account.AccountID
		want     account.Username
		wantErr  bool
	}{
		{
			name:     "generates archived username from normal username",
			username: account.Username("alice"),
			id:       account.AccountID(9223372036854775807),
			want:     account.Username("alice#archived_9223372036854775807"),
			wantErr:  false,
		},
		{
			name:     "keeps archived username within storage limit",
			username: account.Username(strings.Repeat("a", account.MaxUsernameLength)),
			id:       account.AccountID(9223372036854775807),
			want:     account.Username(strings.Repeat("a", account.MaxUsernameLength) + "#archived_9223372036854775807"),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := account.ArchiveUsername(tt.username, tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ArchiveUsername() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ArchiveUsername() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("ArchiveUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}
