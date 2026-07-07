package account

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Username string

func NewUsername(value string) (Username, error) {
	trimmed := strings.TrimSpace(value)
	length := utf8.RuneCountInString(trimmed)

	if length == 0 {
		return "", fmt.Errorf(
			"%w: username cannot be blank",
			ErrUsernameBlank,
		)
	}

	if length < MinUsernameLength || length > MaxUsernameLength {
		return "", fmt.Errorf(
			"%w: username length must be between %d and %d characters",
			ErrUsernameLength,
			MinUsernameLength,
			MaxUsernameLength,
		)
	}

	return Username(trimmed), nil
}

func (u Username) Value() string {
	return string(u)
}

func UsernameFromStorage(value string) (Username, error) {
	trimmed := strings.TrimSpace(value)
	length := utf8.RuneCountInString(trimmed)

	if length == 0 {
		return "", fmt.Errorf(
			"%w: stored username is required",
			ErrUsernameBlank,
		)
	}

	if length > MaxStoredUsernameLength {
		return "", fmt.Errorf(
			"%w: stored username must not exceed %d characters",
			ErrUsernameLength,
			MaxStoredUsernameLength,
		)
	}

	return Username(trimmed), nil
}

func ArchiveUsername(username Username, id ID) (Username, error) {
	value := fmt.Sprintf(archivedUsernameFormat, username.Value(), id)
	length := utf8.RuneCountInString(value)

	if length > MaxStoredUsernameLength {
		return "", fmt.Errorf(
			"%w: stored username must not exceed %d characters",
			ErrUsernameLength,
			MaxStoredUsernameLength,
		)
	}

	return Username(value), nil
}
