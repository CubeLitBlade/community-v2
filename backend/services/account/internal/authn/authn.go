// Package authn provides authentication middleware for extracting and validating
// JWT claims from HTTP requests.
package authn

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/cubelitblade/community-v2/backend/pkg/common/httperr"
	"github.com/cubelitblade/community-v2/backend/pkg/common/jwt"
	"github.com/gin-gonic/gin"
)

const keyPrincipal = "principal"

// Principal represents the authenticated user extracted from a JWT.
type Principal struct {
	ID   int64
	Role string
}

// ParseToken is a Gin middleware that extracts a JWT from the access_token
// cookie, validates it, and sets the Principal in the Gin context.
func ParseToken(parser *jwt.Parser, cookieName string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.With("cookie", cookieName)

		token, err := c.Cookie(cookieName)
		if err != nil {
			log.Debug("cookie not found")
			c.Next()

			return
		}

		claims, err := parser.Parse(token)
		if err != nil {
			switch {
			case errors.Is(err, jwt.ErrInvalidToken):
				log.Debug("token invalid", "error", err)
			case errors.Is(err, jwt.ErrTokenExpired):
				log.Debug("token expired")
			default:
				log.Debug("token parse error", "error", err)
			}

			c.Next()

			return
		}

		id, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			log.Debug("subject parse error", "error", err)
			c.Next()

			return
		}

		c.Set(keyPrincipal, Principal{ID: id, Role: claims.Role})
		c.Next()
	}
}

// MustAuthenticate is a Gin middleware that aborts the request with 401
// Unauthorized if no Principal is present in the context.
func MustAuthenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := GetPrincipal(c)
		if !ok {
			httperr.WriteUnauthorized(c, "Authorization required.")
			c.Abort()

			return
		}

		c.Next()
	}
}

// GetPrincipal retrieves the Principal from the Gin context.
func GetPrincipal(c *gin.Context) (Principal, bool) {
	val, ok := c.Get(keyPrincipal)
	if !ok {
		//nolint:exhaustruct // Returning zero-value is idiomatic and safe for the failure path; caller must check the bool.
		return Principal{}, false
	}

	principal, ok := val.(Principal)
	if !ok {
		//nolint:exhaustruct // Returning zero-value is idiomatic and safe for the failure path; caller must check the bool.
		return Principal{}, false
	}

	return principal, true
}
