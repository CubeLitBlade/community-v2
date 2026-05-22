package account

import (
	"errors"
	"fmt"
)

// Role defines a account's privilege level within the system.
type Role string

// ErrRoleUnknown indicates that a role string does not correspond to a known
// Role.
var ErrRoleUnknown = errors.New("unknown role")

// Available role constants.
const (
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleMember    = "member"
)

// ParseRole converts a string to a Role.
// It returns ErrRoleUnknown if the string is not a valid role.
func ParseRole(value string) (Role, error) {
	role := Role(value)

	if !role.IsValid() {
		return "", fmt.Errorf(
			"%w: role '%s' is unknown",
			ErrRoleUnknown,
			value,
		)
	}

	return role, nil
}

func mustParseRole(value string) Role {
	role, err := ParseRole(value)
	if err != nil {
		panic(err)
	}

	return role
}

// IsValid reports whether r is a recognized Role value.
func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleModerator, RoleMember:
		return true
	default:
		return false
	}
}

func (r Role) String() string {
	return string(r)
}
