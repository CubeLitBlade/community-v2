// Package transport provides HTTP transport adapters for the auth domain.
package transport

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/CubeLitBlade/community-v2/backend/internal/auth"
	"github.com/CubeLitBlade/community-v2/backend/internal/httperr"
)

// Handler handles HTTP requests for authentication.
type Handler struct {
	login  *auth.Login
	logger *slog.Logger
}

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

	g.POST("/", h.Login)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

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

	jwt, err := h.login.Execute(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.logger.Debug(
			"login failed",
			"username", req.Username,
			"error", err,
		)

		httperr.WriteMappedError(c, err, authProblem)

		return
	}

	c.JSON(http.StatusOK, loginResponse{
		Token: jwt,
	})
}
