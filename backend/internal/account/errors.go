package account

import "errors"

// ErrAccountNotFound is returned when the account is not found in database.
var ErrAccountNotFound = errors.New("account not found")
