package account_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/CubeLitBlade/community-v2/backend/services/user/internal/account"
)

const (
	fill              string     = "a"
	testID            account.ID = 1
	maxArchiveID      account.ID = 9223372036854775807
	maxArchivedSuffix            = "#archived_9223372036854775807"
)

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

			got, gotErr := account.ArchiveUsername(
				testCase.username,
				testCase.id,
			)
			assertUsernameResult(
				t,
				"ArchiveUsername",
				got,
				gotErr,
				testCase.want,
				testCase.wantErr,
			)
		})
	}
}

func runUsernameCases(
	t *testing.T,
	funcName string,
	factory func(string) (account.Username, error),
	tests []usernameCase,
) {
	t.Helper()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := factory(testCase.value)
			assertUsernameResult(
				t,
				funcName,
				got,
				gotErr,
				testCase.want,
				testCase.wantErr,
			)
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
	minUsername := strings.Repeat(fill, account.MinUsernameLength)
	maxUsername := strings.Repeat(fill, account.MaxUsernameLength)

	return []usernameCase{
		{
			name:    "accepts minimum length username",
			value:   minUsername,
			want:    account.Username(minUsername),
			wantErr: nil,
		},
		{
			name:    "accepts maximum length username",
			value:   maxUsername,
			want:    account.Username(maxUsername),
			wantErr: nil,
		},
		{
			name:    "rejects username shorter than minimum length",
			value:   fill,
			want:    "",
			wantErr: account.ErrUsernameLength,
		},
		{
			name:    "rejects username longer than maximum length",
			value:   strings.Repeat(fill, account.MaxUsernameLength+1),
			want:    "",
			wantErr: account.ErrUsernameLength,
		},
		{
			name:    "trims surrounding spaces",
			value:   " Alice ",
			want:    "Alice",
			wantErr: nil,
		},
		{
			name:    "rejects blank username",
			value:   "   ",
			want:    "",
			wantErr: account.ErrUsernameBlank,
		},
	}
}

func usernameFromStorageCases() []usernameCase {
	tooLongNormalUsername := strings.Repeat(fill, account.MaxUsernameLength+1)
	tooLongStoredUsername := strings.Repeat(
		fill,
		account.MaxStoredUsernameLength+1,
	)

	return []usernameCase{
		{
			name:    "accepts normal username",
			value:   "bob",
			want:    "bob",
			wantErr: nil,
		},
		{
			name:    "accepts archived username",
			value:   "Alice#archived_123",
			want:    "Alice#archived_123",
			wantErr: nil,
		},
		{
			name:    "accepts username longer than normal limit",
			value:   tooLongNormalUsername,
			want:    account.Username(tooLongNormalUsername),
			wantErr: nil,
		},
		{
			name:    "rejects blank stored username",
			value:   "   ",
			want:    "",
			wantErr: account.ErrUsernameBlank,
		},
		{
			name:    "rejects username longer than storage limit",
			value:   tooLongStoredUsername,
			want:    "",
			wantErr: account.ErrUsernameLength,
		},
	}
}

func archiveUsernameCases() []archiveUsernameCase {
	maxUsername := strings.Repeat(fill, account.MaxUsernameLength)
	tooLongStoredUsername := strings.Repeat(
		fill,
		account.MaxStoredUsernameLength,
	)

	return []archiveUsernameCase{
		{
			name:     "generates archived username from normal username",
			username: "Alice",
			id:       maxArchiveID,
			want:     "Alice" + maxArchivedSuffix,
			wantErr:  nil,
		},
		{
			name:     "keeps archived username within storage limit",
			username: account.Username(maxUsername),
			id:       maxArchiveID,
			want:     account.Username(maxUsername + maxArchivedSuffix),
			wantErr:  nil,
		},
		{
			name:     "rejects archived username longer than storage limit",
			username: account.Username(tooLongStoredUsername),
			id:       testID,
			want:     "",
			wantErr:  account.ErrUsernameLength,
		},
	}
}
