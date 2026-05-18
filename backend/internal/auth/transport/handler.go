// Package transport provides HTTP transport adapters for the auth domain.
package transport

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/CubeLitBlade/community-v2/backend/internal/auth"
	"github.com/CubeLitBlade/community-v2/backend/internal/httperr"
)

const defaultMaxAge = int(time.Hour * 24 / time.Second)

// Handler handles HTTP requests for authentication.
type Handler struct {
	login  *auth.Login
	logger *slog.Logger
}

// Deps holds the dependencies required by the auth Handler.
type Deps struct {
	Login  *auth.Login
	Logger *slog.Logger
}

// NewHandler creates and returns a new AuthHandler.
func NewHandler(deps Deps) *Handler {
	if deps.Logger == nil {
		panic("nil logger")
	}

	if deps.Login == nil {
		panic("nil login")
	}

	return &Handler{
		login:  deps.Login,
		logger: deps.Logger,
	}
}

// RegisterRoutes registers the authentication routes on the given router.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	g := router.Group("/auth")

	g.POST("/login", h.Login)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	ID   int64  `json:"uid"`
	Role string `json:"role"`
}

// Login handles POST /auth/login — authenticates credentials and returns a JWT.
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.WriteInvalidRequest(
			c,
			"Request body must be valid JSON and include"+
				"username and password.",
		)

		return
	}

	session, err := h.login.Execute(
		c.Request.Context(), req.Username, req.Password,
	)
	if err != nil {
		h.logger.Debug(
			"login failed",
			"username", req.Username,
			"error", err,
		)

		httperr.WriteMappedError(c, err, authProblem)

		return
	}

	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    session.Token,
		Path:     "/api",
		MaxAge:   defaultMaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(c.Writer, cookie)

	c.JSON(http.StatusOK, loginResponse{
		ID:   session.ID,
		Role: session.Role,
	})
}
