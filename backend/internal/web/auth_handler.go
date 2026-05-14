package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) RegisterRoutes(router gin.IRouter) {
	auth := router.Group("/auth")

	auth.GET("/csrf", h.csrf)
}

func (h *AuthHandler) csrf(c *gin.Context) {
	c.Header("X-CSRF-Token", csrf.Token(c.Request))
	c.Status(http.StatusNoContent)
}
