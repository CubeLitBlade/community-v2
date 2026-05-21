// Package transport provides HTTP transport adapters for the auth domain.
package transport

import (
	"log/slog"
	"net/http"
	"net/netip"

	"github.com/CubeLitBlade/community-v2/backend/pkg/common/httperr"
	"github.com/CubeLitBlade/community-v2/backend/services/user/internal/auth"
	"github.com/CubeLitBlade/community-v2/backend/services/user/internal/authn"
	"github.com/gin-gonic/gin"
)

const (
	cookiePath     = "/api/"
	cookieHTTPOnly = true
)

// CookieConfig holds the configuration for the auth cookie.
type CookieConfig struct {
	Name     string
	Secure   bool
	SameSite http.SameSite
	MaxAge   int
}

// Handler handles HTTP requests for authentication.
type Handler struct {
	login     *auth.Login
	cookieCfg CookieConfig
	logger    *slog.Logger
}

// Deps holds the dependencies required by the auth Handler.
type Deps struct {
	Login     *auth.Login
	CookieCfg CookieConfig
	Logger    *slog.Logger
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
		login:     deps.Login,
		cookieCfg: deps.CookieCfg,
		logger:    deps.Logger,
	}
}

// RegisterRoutes registers the authentication routes on the given router.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	g := router.Group("/auth")

	g.POST("/", h.Login)
	g.DELETE("/", authn.MustAuthenticate(), h.Logout)
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
		httperr.WriteBadRequest(
			c,
			"Request body must be valid JSON and include username "+
				"and password.",
		)

		return
	}

	ipaddr, err := netip.ParseAddr(c.ClientIP())
	if err != nil {
		httperr.WriteBadRequest(c, "Invalid IP address")

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

	//nolint:gosec // cookie attributes are config-driven; Secure is set
	cookie := &http.Cookie{
		Name:     h.cookieCfg.Name,
		Value:    credentials.Token,
		Path:     cookiePath,
		MaxAge:   h.cookieCfg.MaxAge,
		HttpOnly: cookieHTTPOnly,
		Secure:   h.cookieCfg.Secure,
		SameSite: h.cookieCfg.SameSite,
	}

	http.SetCookie(c.Writer, cookie)
	c.JSON(http.StatusOK, loginResponse{
		ID:   credentials.ID,
		Role: credentials.Role,
	})
}

// Logout handles DELETE /auth/ — erase access token cookie.
func (h *Handler) Logout(c *gin.Context) {
	//nolint:gosec // cookie attributes are config-driven; Secure is set
	cookie := &http.Cookie{
		Name:     h.cookieCfg.Name,
		Value:    "",
		Path:     cookiePath,
		MaxAge:   -1, // delete cookie immediately
		HttpOnly: cookieHTTPOnly,
		Secure:   h.cookieCfg.Secure,
		SameSite: h.cookieCfg.SameSite,
	}

	http.SetCookie(c.Writer, cookie)
	c.Status(http.StatusNoContent)
}
