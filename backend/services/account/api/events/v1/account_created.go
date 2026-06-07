package v1

const (
	// TopicAccountCreated is the topic for account created events.
	TopicAccountCreated = "account.created.v1"
	// EventTypeAccountCreated is Cloud Event type for account created events.
	EventTypeAccountCreated = "io.github.cubelitblade.account.created.v1"
)

// AccountCreatedEventPayload is the event payload for when a new account is created.
type AccountCreatedEventPayload struct {
	AccountID string `json:"accountId"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}
