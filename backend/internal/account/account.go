package account

import (
	"errors"
	"fmt"
	"net/netip"
	"time"
)

// ErrInvalidAccountID is returned when the account ID is invalid.
var ErrInvalidAccountID = errors.New("invalid account id")

// ID represents the unique identifier of an Account.
type ID int64

// Account represents a user account with authentication and profile data.
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

// Register creates a new Account with the given ID, username, and password.
func Register(
	id int64,
	username, password string,
	now time.Time,
) (Account, error) {
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

// RecordLogin updates last login time and last login ip
func (a *Account) RecordLogin(at time.Time, addr netip.Addr) {
	a.lastLogin = &LastLogin{
		Time:   at,
		IPAddr: addr,
	}
}

func newID(value int64) (ID, error) {
	if value <= 0 {
		return 0, fmt.Errorf(
			"%w: invalid account id '%d'",
			ErrInvalidAccountID,
			value,
		)
	}

	return ID(value), nil
}

// ID returns the ID of the account.
func (a *Account) ID() ID {
	return a.id
}

// Username returns the username of the account.
func (a *Account) Username() string {
	return a.username.Value()
}

// DisplayName returns the display name of the account.
func (a *Account) DisplayName() string {
	return a.displayName
}

// Role returns the role of the account.
func (a *Account) Role() Role {
	return a.role
}

// Status returns the status of the account.
func (a *Account) Status() Status {
	return a.status
}

// Snapshot represents a point-in-time copy of an Account's state,
// including audit information and authentication details.
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

// Snapshot returns a Snapshot containing the current state of the account,
// including audit information and last login details.
func (a *Account) Snapshot() Snapshot {
	var (
		lastLoginAt *time.Time
		lastLoginIP *netip.Addr
	)

	if a.lastLogin != nil {
		lastLoginAt = new(a.lastLogin.Time)
		lastLoginIP = new(a.lastLogin.IPAddr)
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

// NewAccountFromSnapshot reconstructs account from snapshot.
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

// VerifyPassword checks whether the given password matches the account's password hash.
func (a *Account) VerifyPassword(password string) bool {
	return a.passwordHash.matches(password)
}
