package account

import (
	"fmt"
	"net/netip"
	"time"
)

type ID int64

type Account struct {
	createdAt, updatedAt   time.Time
	lastLogin              *LastLogin
	username               Username
	passwordHash           PasswordHash
	displayName            string
	role                   Role
	status                 Status
	id                     ID
	passwordChangeRequired bool
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
		id:                     accountID,
		username:               accountUsername,
		passwordHash:           accountPasswordHash,
		passwordChangeRequired: false,
		displayName:            accountUsername.Value(),
		role:                   RoleMember,
		status:                 StatusActive,
		createdAt:              now,
		updatedAt:              now,
		lastLogin:              nil,
	}, nil
}

func (a *Account) RecordLogin(at time.Time, addr netip.Addr) {
	a.lastLogin = &LastLogin{
		Time:   at,
		IPAddr: addr,
	}
}

func newID(value int64) (ID, error) {
	if value <= 0 {
		return 0, fmt.Errorf("%w: invalid value: %d", ErrInvalidAccountID, value)
	}

	return ID(value), nil
}

func (a *Account) ID() int64 {
	return int64(a.id)
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
	Role                   string
	Status                 string
	ID                     int64
	PasswordChangeRequired bool
}

func (a *Account) Snapshot() Snapshot {
	var (
		lastLoginAt *time.Time
		lastLoginIP *netip.Addr
	)

	if a.lastLogin != nil {
		t := a.lastLogin.Time
		lastLoginAt = &t
		ip := a.lastLogin.IPAddr
		lastLoginIP = &ip
	}

	return Snapshot{
		ID:                     int64(a.id),
		Username:               a.username.Value(),
		PasswordHash:           a.passwordHash.value,
		PasswordChangeRequired: a.passwordChangeRequired,
		DisplayName:            a.displayName,
		Role:                   a.role.String(),
		Status:                 a.status.String(),
		CreatedAt:              a.createdAt,
		UpdatedAt:              a.updatedAt,
		LastLoginAt:            lastLoginAt,
		LastLoginIP:            lastLoginIP,
	}
}

func NewAccountFromSnapshot(snap Snapshot) Account {
	var lastLogin *LastLogin
	if snap.LastLoginAt != nil && snap.LastLoginIP != nil {
		lastLogin = NewLastLogin(*snap.LastLoginAt, *snap.LastLoginIP)
	}

	return Account{
		id:                     ID(snap.ID),
		username:               Username(snap.Username),
		passwordHash:           PasswordHash{value: snap.PasswordHash},
		passwordChangeRequired: snap.PasswordChangeRequired,
		displayName:            snap.DisplayName,
		role:                   mustParseRole(snap.Role),
		status:                 mustParseStatus(snap.Status),
		createdAt:              snap.CreatedAt,
		updatedAt:              snap.UpdatedAt,
		lastLogin:              lastLogin,
	}
}

func (a *Account) VerifyPassword(password string) bool {
	return a.passwordHash.Matches(password)
}
