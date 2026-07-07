package v1

type AccountCreated struct {
	AccountId string `json:"accountId"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}
