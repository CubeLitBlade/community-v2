package account

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	MinUsernameLength       = 3
	MaxUsernameLength       = 20
	MaxStoredUsernameLength = 50

	archivedUsernameFormat = "%s#archived_%d"
)

type Username string

func NewUsername(value string) (Username, error) {
	value = strings.TrimSpace(value)
	length := utf8.RuneCountInString(value)

	if length < MinUsernameLength || length > MaxUsernameLength {
		return "", fmt.Errorf(
			"%w: username length must be between %d and %d characters",
			ErrInvalidUsername,
			MinUsernameLength,
			MaxUsernameLength,
		)
	}

	return Username(value), nil
}

func (u Username) String() string {
	return string(u)
}

func UsernameFromStorage(value string) (Username, error) {
	value = strings.TrimSpace(value)
	length := utf8.RuneCountInString(value)

	if value == "" {
		return "", fmt.Errorf(
			"%w: stored username is required",
			ErrInvalidUsername,
		)
	}

	if length > MaxStoredUsernameLength {
		return "", fmt.Errorf(
			"%w: stored username must not exceed %d characters",
			ErrInvalidUsername,
			MaxStoredUsernameLength,
		)
	}

	return Username(value), nil
}

func ArchiveUsername(username Username, id AccountID) (Username, error) {
	value := fmt.Sprintf(archivedUsernameFormat, username.String(), id)
	length := utf8.RuneCountInString(value)

	if length > MaxStoredUsernameLength {
		return "", fmt.Errorf(
			"%w: stored username must not exceed %d characters",
			ErrInvalidUsername,
			MaxStoredUsernameLength,
		)
	}
	return Username(value), nil
}
