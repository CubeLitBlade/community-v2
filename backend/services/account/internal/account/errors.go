package account

import "errors"

// Sentinel errors for the account domain.
var (
	ErrAccountNotFound       = errors.New("account not found")
	ErrNilAccount            = errors.New("account is nil")
	ErrUsernameAlreadyExists = errors.New("username already exists")

	ErrInvalidAccountID   = errors.New("invalid account id")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrPasswordEmpty     = errors.New("password cannot be empty")
	ErrPasswordTooShort  = errors.New("password too short")
	ErrPasswordTooLong   = errors.New("password too long")
	ErrPasswordHashEmpty = errors.New("password hash cannot be empty")

	ErrUsernameLength = errors.New("invalid username length")
	ErrUsernameBlank  = errors.New("username is blank")

	ErrRoleUnknown   = errors.New("unknown role")
	ErrStatusUnknown = errors.New("unknown status")
)
