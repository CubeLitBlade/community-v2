// Package transport provides HTTP transport adapters for the auth domain.
package transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
)

// Handler handles HTTP requests for authentication.
type Handler struct{}

// NewHandler creates and returns a new AuthHandler.
func NewHandler() *Handler {
	return &Handler{}
}

// RegisterRoutes registers the authentication routes on the given router.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	auth := router.Group("/auth")

	auth.GET("/csrf", h.csrf)
}

func (*Handler) csrf(c *gin.Context) {
	c.Header("X-CSRF-Token", csrf.Token(c.Request))
	c.Status(http.StatusNoContent)
}
