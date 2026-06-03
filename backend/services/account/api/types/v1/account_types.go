// Package v1 defines the request and response types for the account service's v1 API.
package v1

// CreateAccountRequest is the request body for creating a new account.
type CreateAccountRequest struct {
	Username string `binding:"required" json:"username"`
	Password string `binding:"required,min=15" json:"password"`
}

// Profile is the response body for an account profile.
type Profile struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}
