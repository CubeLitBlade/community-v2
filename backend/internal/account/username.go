package account

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var (
	ErrUsernameLength = errors.New("invalid username length")
	ErrUsernameBlank  = errors.New("username is blank")
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

	if value == "" {
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

	return Username(value), nil
}

func (u Username) Value() string {
	return string(u)
}

func UsernameFromStorage(value string) (Username, error) {
	value = strings.TrimSpace(value)
	length := utf8.RuneCountInString(value)

	if value == "" {
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

	return Username(value), nil
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
