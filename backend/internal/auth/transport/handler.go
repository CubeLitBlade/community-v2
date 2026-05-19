// Package transport provides HTTP transport adapters for the auth domain.
package transport

import (
	"log/slog"
	"net/http"
	"net/netip"

	"github.com/gin-gonic/gin"

	"github.com/CubeLitBlade/community-v2/backend/internal/auth"
	"github.com/CubeLitBlade/community-v2/backend/internal/httperr"
)

const (
	cookieName     = "access_token"
	cookiePath     = "/api/"
	cookieMaxAge   = 15 * 60
	cookieHTTPOnly = true
	cookieSecure   = true
	cookieSameSite = http.SameSiteLaxMode
)

// Handler handles HTTP requests for authentication.
type Handler struct {
	login   *auth.Login
	logger  *slog.Logger
	authnMW func(c *gin.Context)
}

// Deps holds the dependencies required by the auth Handler.
type Deps struct {
	Login   *auth.Login
	Logger  *slog.Logger
	AuthnMW func(c *gin.Context)
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
		login:   deps.Login,
		logger:  deps.Logger,
		authnMW: deps.AuthnMW,
	}
}

// RegisterRoutes registers the authentication routes on the given router.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	g := router.Group("/auth")

	g.POST("/", h.Login)
	g.DELETE("/", h.authnMW, h.Logout)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	ID   int64  `json:"uid"`
	Role string `json:"role"`
}

// Login handles POST /auth/ — authenticates credentials and returns a JWT.
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.WriteInvalidRequest(
			c,
			"Request body must be valid JSON and include username "+
				"and password.",
		)

		return
	}

	ipaddr, err := netip.ParseAddr(c.ClientIP())
	if err != nil {
		httperr.WriteInvalidRequest(c, "Invalid IP address")

		return
	}

	credentials, err := h.login.Execute(
		c.Request.Context(), req.Username, req.Password, ipaddr,
	)
	if err != nil {
		h.logger.Debug(
			"login failed", "username", req.Username, "error", err,
		)

		httperr.WriteMappedError(c, err, authProblem)

		return
	}

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    credentials.Token,
		Path:     cookiePath,
		MaxAge:   cookieMaxAge,
		HttpOnly: cookieHTTPOnly,
		Secure:   cookieSecure,
		SameSite: cookieSameSite,
	}

	http.SetCookie(c.Writer, cookie)

	c.JSON(http.StatusOK, loginResponse{
		ID:   credentials.ID,
		Role: credentials.Role,
	})
}

// Logout handles DELETE /auth/ — erase access token cookie.
func (h *Handler) Logout(c *gin.Context) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     cookiePath,
		MaxAge:   -1, // delete cookie immediately
		HttpOnly: cookieHTTPOnly,
		Secure:   cookieSecure,
		SameSite: cookieSameSite,
	}

	http.SetCookie(c.Writer, cookie)
	c.Status(http.StatusNoContent)
}
