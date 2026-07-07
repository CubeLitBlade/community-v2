package account

import "errors"

var (
	ErrAccountNotFound       = errors.New("account not found")
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
