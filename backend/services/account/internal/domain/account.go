package account

import (
	"fmt"
	"net/netip"
	"time"
)

type ID int64

type Account struct {
	CreatedAt              time.Time
	UpdatedAt              time.Time
	LastLoginAt            *time.Time
	LastLoginIP            *netip.Addr
	Username               Username
	PasswordHash           PasswordHash
	DisplayName            string
	Role                   Role
	Status                 Status
	ID                     ID
	PasswordChangeRequired bool
}

func NewAccount(id int64, username, password string, now time.Time) (Account, error) {
	accountID, err := newID(id)
	if err != nil {
		return Account{}, err
	}

	accountUsername, err := NewUsername(username)
	if err != nil {
		return Account{}, err
	}

	accountPasswordHash, err := HashPassword(password)
	if err != nil {
		return Account{}, err
	}

	return Account{
		ID:                     accountID,
		Username:               accountUsername,
		PasswordHash:           accountPasswordHash,
		PasswordChangeRequired: false,
		DisplayName:            accountUsername.Value(),
		Role:                   RoleMember,
		Status:                 StatusActive,
		CreatedAt:              now,
		UpdatedAt:              now,
		LastLoginAt:            nil,
		LastLoginIP:            nil,
	}, nil
}

func (a *Account) RecordLogin(at time.Time, addr netip.Addr) {
	a.LastLoginAt = &at
	a.LastLoginIP = &addr
}

func newID(value int64) (ID, error) {
	if value <= 0 {
		return 0, fmt.Errorf("%w: invalid value: %d", ErrInvalidAccountID, value)
	}

	return ID(value), nil
}

func (a *Account) VerifyPassword(password string) bool {
	return a.PasswordHash.Matches(password)
}
