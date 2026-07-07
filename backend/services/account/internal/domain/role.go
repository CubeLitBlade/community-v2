package account

import (
	"fmt"
)

type Role string

func ParseRole(value string) (Role, error) {
	role := Role(value)

	if !role.IsValid() {
		return "", fmt.Errorf("%w: role '%s' is unknown", ErrRoleUnknown, value)
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
