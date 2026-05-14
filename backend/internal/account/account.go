package account

import (
	"errors"
	"fmt"
	"net/netip"
	"time"
)

var ErrInvalidAccountID = errors.New("invalid account id")

type ID int64

type Account struct {
	audit                  Audit
	username               Username
	passwordHash           PasswordHash
	displayName            string
	role                   Role
	status                 Status
	id                     ID
	passwordChangeRequired bool
}

func Register(id ID, username, password string, now time.Time) (Account, error) {
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

func (a *Account) ID() ID {
	return a.id
}

func (a *Account) Username() string {
	return a.username.Value()
}

func (a *Account) DisplayName() string {
	return a.displayName
}

func (a *Account) Role() Role {
	return a.role
}

func (a *Account) Status() Status {
	return a.status
}

type Snapshot struct {
	CreatedAt              time.Time
	UpdatedAt              time.Time
	LastLoginAt            *time.Time
	LastLoginIP            *netip.Addr
	Username               string
	PasswordHash           string
	DisplayName            string
	Role                   Role
	Status                 Status
	ID                     ID
	PasswordChangeRequired bool
}

func (a *Account) Snapshot() Snapshot {
	var (
		lastLoginAt *time.Time
		lastLoginIP *netip.Addr
	)

	if a.audit.lastLogin != nil {
		at := a.audit.lastLogin.at
		ip := a.audit.lastLogin.ip

		lastLoginAt = &at
		lastLoginIP = &ip
	}

	return Snapshot{
		ID:                     a.id,
		Username:               a.username.Value(),
		PasswordHash:           a.passwordHash.value,
		PasswordChangeRequired: a.passwordChangeRequired,
		DisplayName:            a.displayName,
		Role:                   a.role,
		Status:                 a.status,
		CreatedAt:              a.audit.createdAt,
		UpdatedAt:              a.audit.updatedAt,
		LastLoginAt:            lastLoginAt,
		LastLoginIP:            lastLoginIP,
	}
}
