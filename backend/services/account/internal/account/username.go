package account

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var (
	// ErrUsernameLength indicates that a username's length is out of the
	// allowed bounds.
	ErrUsernameLength = errors.New("invalid username length")

	// ErrUsernameBlank indicates that a username is empty or consists only of
	// whitespace.
	ErrUsernameBlank = errors.New("username is blank")
)

// Username length constraints.
const (
	MinUsernameLength       = 3
	MaxUsernameLength       = 20
	MaxStoredUsernameLength = 50

	archivedUsernameFormat = "%s#archived_%d"
)

// Username represents a validated account identifier.
type Username string

// emptyUsername is a named constant replacing the repeated "" string literal.
const emptyUsername Username = ""

// NewUsername creates a new Username after validating its length and content.
func NewUsername(value string) (Username, error) {
	trimmed := strings.TrimSpace(value)
	length := utf8.RuneCountInString(trimmed)

	if length == 0 {
		return emptyUsername, fmt.Errorf(
			"%w: username cannot be blank",
			ErrUsernameBlank,
		)
	}

	if length < MinUsernameLength || length > MaxUsernameLength {
		return emptyUsername, fmt.Errorf(
			"%w: username length must be between %d and %d characters",
			ErrUsernameLength,
			MinUsernameLength,
			MaxUsernameLength,
		)
	}

	return Username(trimmed), nil
}

// Value returns the underlying string value of the Username.
func (u Username) Value() string {
	return string(u)
}

// UsernameFromStorage reconstructs a Username from a storage layer value.
// It applies different validation rules, such as allowing longer archived
// names.
func UsernameFromStorage(value string) (Username, error) {
	trimmed := strings.TrimSpace(value)
	length := utf8.RuneCountInString(trimmed)

	if length == 0 {
		return emptyUsername, fmt.Errorf(
			"%w: stored username is required",
			ErrUsernameBlank,
		)
	}

	if length > MaxStoredUsernameLength {
		return emptyUsername, fmt.Errorf(
			"%w: stored username must not exceed %d characters",
			ErrUsernameLength,
			MaxStoredUsernameLength,
		)
	}

	return Username(trimmed), nil
}

// ArchiveUsername generates an archived version of a Username by appending the
// account ID.
func ArchiveUsername(username Username, id ID) (Username, error) {
	value := fmt.Sprintf(archivedUsernameFormat, username.Value(), id)
	length := utf8.RuneCountInString(value)

	if length > MaxStoredUsernameLength {
		return emptyUsername, fmt.Errorf(
			"%w: stored username must not exceed %d characters",
			ErrUsernameLength,
			MaxStoredUsernameLength,
		)
	}

	return Username(value), nil
}
