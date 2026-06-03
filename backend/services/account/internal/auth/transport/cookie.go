// Package transport provides HTTP transport helpers for the auth domain.
package transport

import (
	"net/http"
	"time"
)

// Cookie name and path constants.
const (
	// CookieNameAccessToken is the name of the access token cookie.
	CookieNameAccessToken = "access_token"
	// CookiePath is the path for which the cookie is valid.
	CookiePath = "/api/"
)

// CookiePolicy provides the security policy for cookies.
type CookiePolicy interface {
	Secure() bool
	SameSite() http.SameSite
}

// WriteCookie creates a new http.Cookie with the given parameters and the policy's security settings.
func WriteCookie(name string, value string, ttl time.Duration, policy CookiePolicy) *http.Cookie {
	//nolint:gosec // Secure and SameSite is provided by CookiePolicy.
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     CookiePath,
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   policy.Secure(),
		SameSite: policy.SameSite(),
	}
}
