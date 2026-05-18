// Package authn provides authorization middleware for extracting and validating
// JWT claims from HTTP requests.
package authn

import (
	"errors"
	"log"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/CubeLitBlade/community-v2/backend/internal/httperr"
	"github.com/CubeLitBlade/community-v2/backend/internal/jwt"
)

const keyPrincipal = "principal"

// Principal represents the authenticated user extracted from a JWT.
type Principal struct {
	ID   int64
	Role string
}

// Middleware returns a Gin middleware that validates the access_token cookie
// and injects the Principal into the request context.
func Middleware(jwtKey string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			logger.Warn("access_token cookie not found")
			httperr.WriteUnauthorized(c, "Missing access_token")
			c.Abort()

			return
		}

		claims, err := jwt.Parse(token, jwtKey)
		if err != nil {
			if errors.Is(err, jwt.ErrInvalidToken) {
				httperr.WriteUnauthorized(c, "Invalid access_token")
			} else if errors.Is(err, jwt.ErrTokenExpired) {
				httperr.WriteUnauthorized(c, "Token is expired")
			}

			c.Abort()

			return
		}

		id, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			httperr.WriteUnauthorized(c, "error parsing id")
			c.Abort()

			return
		}

		log.Print(id, claims.Role)

		c.Set(keyPrincipal, Principal{
			ID:   id,
			Role: claims.Role,
		})

		c.Next()
	}
}

// GetPrincipal retrieves the Principal from the Gin context.
func GetPrincipal(c *gin.Context) (Principal, bool) {
	p, ok := c.Get(keyPrincipal)
	if !ok {
		return Principal{ID: 0, Role: ""}, false
	}

	principal, ok := p.(Principal)
	if !ok {
		return Principal{ID: 0, Role: ""}, false
	}

	return principal, true
}
