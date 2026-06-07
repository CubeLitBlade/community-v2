// Package v1 defines the request and response types for the authz service's v1 API.
package v1

// AuthorizationRequest is the request body for checking a user's permission.
type AuthorizationRequest struct {
	User     string `binding:"required" json:"user"`
	Relation string `binding:"required" json:"relation"`
	Object   string `binding:"required" json:"object"`
}

// Decision is the response body representing an authorization check result.
type Decision struct {
	Allowed bool `json:"allowed"`
}
