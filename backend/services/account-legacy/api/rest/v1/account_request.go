// Package v1 defines the request and response types for the account service's v1 API.
package v1

// CreateAccountRequest is the request body for creating a new account.
type CreateAccountRequest struct {
	Username string `binding:"required" json:"username"`
	Password string `binding:"required,min=15" json:"password"`
}
