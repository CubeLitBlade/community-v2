package account

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordEmpty     = errors.New("password cannot be empty")
	ErrPasswordTooShort  = errors.New("password too short")
	ErrPasswordTooLong   = errors.New("password too long")
	ErrPasswordHashEmpty = errors.New("password hash cannot be empty")
)

const (
	MinPasswordLength = 15
	MaxPasswordBytes  = 72

	passwordCost = bcrypt.DefaultCost
)

type PasswordHash struct {
	value string
}

func HashPassword(raw string) (PasswordHash, error) {
	err := ValidatePassword(raw)

	if err != nil {
		return PasswordHash{}, err
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(raw), passwordCost)

	if err != nil {
		return PasswordHash{}, err
	}

	return PasswordHash{value: string(hashedBytes)}, nil
}

func PasswordMatches(hash PasswordHash, raw string) bool {
	if raw == "" || len([]byte(raw)) > MaxPasswordBytes {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(hash.value), []byte(raw)) == nil
}

func ValidatePassword(raw string) error {
	if raw == "" {
		return fmt.Errorf(
			"%w: password cannot be empty",
			ErrPasswordEmpty,
		)
	}

	length := utf8.RuneCountInString(raw)
	if length < MinPasswordLength {
		return fmt.Errorf(
			"%w: password should be at least %d characters",
			ErrPasswordTooShort,
			MinPasswordLength,
		)
	}

	if len([]byte(raw)) > MaxPasswordBytes {
		return fmt.Errorf(
			"%w: password must not exceed %d bytes",
			ErrPasswordTooLong,
			MaxPasswordBytes,
		)
	}

	return nil
}

func PasswordHashFromStorage(hash string) (PasswordHash, error) {
	if hash == "" {
		return PasswordHash{}, fmt.Errorf(
			"%w: stored password hash is required",
			ErrPasswordHashEmpty,
		)
	}

	return PasswordHash{value: hash}, nil
}

func (p PasswordHash) Value() string {
	return p.value
}

func (p PasswordHash) String() string {
	// Avoid plaintext during debugging
	return "[redacted]"
}
