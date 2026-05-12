package account

type AccountID int64

type Account struct {
	ID           AccountID
	Username     Username
	PasswordHash PasswordHash
}
