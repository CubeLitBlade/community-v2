package account

import (
	"errors"
	"fmt"
)

// Status defines the current state of an account within the system.
type Status string

// ErrStatusUnknown indicates that a status string does not correspond to a
// known Status.
var ErrStatusUnknown = errors.New("unknown status")

// Available status constants.
const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusArchived  Status = "archived"
)

// ParseStatus converts a string to a Status.
// It returns ErrStatusUnknown if the string is not a valid status.
func ParseStatus(value string) (Status, error) {
	status := Status(value)

	if !status.IsValid() {
		return "", fmt.Errorf(
			"%w: status '%s' is unknown",
			ErrStatusUnknown,
			value,
		)
	}

	return status, nil
}

// IsValid reports whether s is a recognized Status value.
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusSuspended, StatusArchived:
		return true
	default:
		return false
	}
}

func (s Status) String() string {
	return string(s)
}

// CanLogin reports whether an account with this status is allowed to log in.
func (s Status) CanLogin() bool {
	return s == StatusActive
}

// CanBeSuspended reports whether an account with this status can be suspended.
func (s Status) CanBeSuspended() bool {
	return s == StatusActive
}

// CanBeRestored reports whether an account with this status can be restored to
// active.
func (s Status) CanBeRestored() bool {
	return s == StatusSuspended
}
