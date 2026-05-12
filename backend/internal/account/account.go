package account

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidAccountID = errors.New("invalid account id")
)

type ID int64

type Account struct {
	id                     ID
	username               Username
	passwordHash           PasswordHash
	passwordChangeRequired bool
	displayName            string
	role                   Role
	status                 Status
	audit                  Audit
}

func Register(id ID, username string, password string, now time.Time) (Account, error) {
	if id <= 0 {
		return Account{}, fmt.Errorf(
			"%w: invalid account id '%d'",
			ErrInvalidAccountID,
			id,
		)
	}

	u, err := NewUsername(username)
	if err != nil {
		return Account{}, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return Account{}, err
	}

	return Account{
		id:                     id,
		username:               u,
		passwordHash:           hash,
		passwordChangeRequired: false,
		displayName:            u.Value(),
		role:                   RoleMember,
		status:                 StatusActive,
		audit:                  NewAudit(now),
	}, nil
}
