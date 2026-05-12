package account

type ID int64

type Account struct {
	id                     ID
	username               Username
	passwordHash           PasswordHash
	passwordChangeRequired bool
	displayName            string
	role                   Role
	status                 Status
	audit                  Audit
}
