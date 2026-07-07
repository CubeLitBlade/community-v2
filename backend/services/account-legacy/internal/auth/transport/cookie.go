// Package transport provides HTTP transport helpers for the auth domain.
package transport

import (
	"net/http"
	"time"

	v1 "github.com/cubelitblade/community-v2/backend/services/account-legacy/api/rest/v1"
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
		Path:     v1.CookiePath,
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   policy.Secure(),
		SameSite: policy.SameSite(),
	}
}
