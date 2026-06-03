// Package contracts defines the shared types and interfaces used across the account service.
package contracts

import (
	"time"

	jwt2 "github.com/golang-jwt/jwt/v5"
)

// AccessTokenClaims defines the JWT claims for an access token, including the user's role.
type AccessTokenClaims struct {
	jwt2.RegisteredClaims

	Role string `json:"role"`
}

// AuthenticatedAccount carries the minimal account information needed after a successful authentication.
type AuthenticatedAccount struct {
	ID   int64
	Role string
}

// TTLProvider provides the access token time-to-live.
type TTLProvider interface {
	AccessTokenTTL() time.Duration
}
