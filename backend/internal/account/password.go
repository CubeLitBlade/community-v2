package account

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrPasswordEmpty indicates that the password string is empty.
	ErrPasswordEmpty = errors.New("password cannot be empty")

	// ErrPasswordTooShort indicates that the password is shorter than
	// MinPasswordLength.
	ErrPasswordTooShort = errors.New("password too short")

	// ErrPasswordTooLong indicates that the password exceeds MaxPasswordBytes.
	ErrPasswordTooLong = errors.New("password too long")

	// ErrPasswordHashEmpty indicates that a stored password hash is empty.
	ErrPasswordHashEmpty = errors.New("password hash cannot be empty")
)

const (
	// MinPasswordLength is the minimal number of characters a password should
	// have.
	MinPasswordLength = 15

	// MaxPasswordBytes is the maximum length in bytes for bcrypt passwords.
	MaxPasswordBytes = 72

	passwordCost = bcrypt.DefaultCost
	emptyString  = ""
)

// PasswordHash holds a bcrypt-hashed password.
type PasswordHash struct {
	value string
}

// HashPassword hashes a plaintext password and returns a PasswordHash.
// Returns an error if the password is invalid or hashing fails.
func HashPassword(raw string) (PasswordHash, error) {
	err := ValidatePassword(raw)
	if err != nil {
		return PasswordHash{}, err
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(raw), passwordCost)
	if err != nil {
		return PasswordHash{}, fmt.Errorf("hash password: %w", err)
	}

	return PasswordHash{value: string(hashedBytes)}, nil
}

// matches returns true if the plaintext password matches the hashed
// password.
func (p PasswordHash) matches(raw string) bool {
	if raw == emptyString || len([]byte(raw)) > MaxPasswordBytes {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(p.value), []byte(raw)) == nil
}

// ValidatePassword checks if a password meets security requirements.
func ValidatePassword(raw string) error {
	if raw == emptyString {
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

// PasswordHashFromStorage constructs a PasswordHash from a stored hash string.
func PasswordHashFromStorage(hash string) (PasswordHash, error) {
	if hash == emptyString {
		return PasswordHash{}, fmt.Errorf(
			"%w: stored password hash is required",
			ErrPasswordHashEmpty,
		)
	}

	return PasswordHash{value: hash}, nil
}

// Value returns the bcrypt string value of the password hash.
func (p PasswordHash) Value() string {
	return p.value
}

// String returns a redacted string to avoid exposing the password hash.
func (PasswordHash) String() string {
	return "[redacted]"
}
