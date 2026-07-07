package account

import (
	"fmt"
)

type Status string

func ParseStatus(value string) (Status, error) {
	status := Status(value)

	if !status.IsValid() {
		return "", fmt.Errorf("%w: status '%s' is unknown", ErrStatusUnknown, value)
	}

	return status, nil
}

func MustParseStatus(value string) Status {
	status, err := ParseStatus(value)
	if err != nil {
		panic(err)
	}

	return status
}

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

func (s Status) CanLogin() bool {
	return s == StatusActive
}

func (s Status) CanBeSuspended() bool {
	return s == StatusActive
}

func (s Status) CanBeRestored() bool {
	return s == StatusSuspended
}
