package auth

type Claims struct {
	Role      string `json:"role"`
	AccountID int64  `json:"accountId"`
}
