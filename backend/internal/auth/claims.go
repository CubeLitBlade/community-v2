package auth

// Claims represents the JWT claims for an authenticated user.
type Claims struct {
	Role      string `json:"role"`
	AccountID int64  `json:"accountId"`
}
