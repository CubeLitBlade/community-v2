// Package v1 defines the event types and topic constants for the account service's v1 API.
package v1

// AccountCreatedEvent is the event payload for when a new account is created.
type AccountCreatedEvent struct {
	AccountID string `json:"accountId"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}
