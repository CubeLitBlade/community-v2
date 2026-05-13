package account

import (
	"errors"
	"fmt"
)

type Role string

var (
	ErrRoleUnknown = errors.New("unknown role")
)

const (
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleMember    = "member"
)

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
