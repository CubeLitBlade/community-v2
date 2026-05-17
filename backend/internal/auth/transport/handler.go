// Package transport provides HTTP transport adapters for the auth domain.
package transport

import (
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for authentication.
type Handler struct{}

// NewHandler creates and returns a new AuthHandler.
func NewHandler() *Handler {
	return &Handler{}
}

// RegisterRoutes registers the authentication routes on the given router.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
}
