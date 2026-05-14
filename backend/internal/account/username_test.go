package account_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
)

const maxArchiveID account.ID = 9223372036854775807

type usernameCase struct {
	wantErr error
	name    string
	value   string
	want    account.Username
}

type archiveUsernameCase struct {
	wantErr  error
	name     string
	username account.Username
	want     account.Username
	id       account.ID
}

func TestNewUsername(t *testing.T) {
	t.Parallel()

	runUsernameCases(t, "NewUsername", account.NewUsername, newUsernameCases())
}

func TestUsernameFromStorage(t *testing.T) {
	t.Parallel()

	runUsernameCases(
		t,
		"UsernameFromStorage",
		account.UsernameFromStorage,
		usernameFromStorageCases(),
	)
}

func TestArchiveUsername(t *testing.T) {
	t.Parallel()

	for _, testCase := range archiveUsernameCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := account.ArchiveUsername(testCase.username, testCase.id)
			assertUsernameResult(t, "ArchiveUsername", got, gotErr, testCase.want, testCase.wantErr)
		})
	}
}

func runUsernameCases(
	t *testing.T,
	funcName string,
	fn func(string) (account.Username, error),
	tests []usernameCase,
) {
	t.Helper()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := fn(testCase.value)
			assertUsernameResult(t, funcName, got, gotErr, testCase.want, testCase.wantErr)
		})
	}
}

func assertUsernameResult(
	t *testing.T,
	funcName string,
	got account.Username,
	gotErr error,
	want account.Username,
	wantErr error,
) {
	t.Helper()

	if wantErr != nil {
		assertUsernameError(t, funcName, gotErr, wantErr)

		return
	}

	if gotErr != nil {
		t.Fatalf("%s() failed: %v", funcName, gotErr)
	}

	if got != want {
		t.Errorf("%s() = %v, want %v", funcName, got, want)
	}
}

func assertUsernameError(t *testing.T, funcName string, gotErr, wantErr error) {
	t.Helper()

	if gotErr == nil {
		t.Fatalf("%s() succeeded unexpectedly", funcName)
	}

	if !errors.Is(gotErr, wantErr) {
		t.Errorf("%s() error = %v, want %v", funcName, gotErr, wantErr)
	}
}

func newUsernameCases() []usernameCase {
	minUsername := strings.Repeat("a", account.MinUsernameLength)
	maxUsername := strings.Repeat("a", account.MaxUsernameLength)

	return []usernameCase{
		{
			name:  "accepts minimum length username",
			value: minUsername,
			want:  account.Username(minUsername),
		},
		{
			name:  "accepts maximum length username",
			value: maxUsername,
			want:  account.Username(maxUsername),
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
}

func usernameFromStorageCases() []usernameCase {
	tooLongNormalUsername := strings.Repeat("a", account.MaxUsernameLength+1)
	tooLongStoredUsername := strings.Repeat("a", account.MaxStoredUsernameLength+1)

	return []usernameCase{
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
			value: tooLongNormalUsername,
			want:  account.Username(tooLongNormalUsername),
		},
		{
			name:    "rejects blank stored username",
			value:   "   ",
			wantErr: account.ErrUsernameBlank,
		},
		{
			name:    "rejects username longer than storage limit",
			value:   tooLongStoredUsername,
			wantErr: account.ErrUsernameLength,
		},
	}
}

func archiveUsernameCases() []archiveUsernameCase {
	maxUsername := strings.Repeat("a", account.MaxUsernameLength)
	tooLongStoredUsername := strings.Repeat("a", account.MaxStoredUsernameLength)

	return []archiveUsernameCase{
		{
			name:     "generates archived username from normal username",
			username: account.Username("alice"),
			id:       maxArchiveID,
			want:     account.Username("alice#archived_9223372036854775807"),
		},
		{
			name:     "keeps archived username within storage limit",
			username: account.Username(maxUsername),
			id:       maxArchiveID,
			want:     account.Username(maxUsername + "#archived_9223372036854775807"),
		},
		{
			name:     "rejects archived username longer than storage limit",
			username: account.Username(tooLongStoredUsername),
			id:       account.ID(1),
			wantErr:  account.ErrUsernameLength,
		},
	}
}
