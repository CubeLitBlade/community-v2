package account

import "golang.org/x/crypto/bcrypt"

const (
	MinPasswordLength = 15
	MaxPasswordBytes  = 72

	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleMember    = "member"

	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusArchived  Status = "archived"

	MinUsernameLength       = 3
	MaxUsernameLength       = 20
	MaxStoredUsernameLength = 50

	passwordCost           = bcrypt.DefaultCost
	archivedUsernameFormat = "%s#archived_%d"
)
