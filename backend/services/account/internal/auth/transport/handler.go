// Package transport provides HTTP transport adapters for the auth domain.
package transport

import (
	"log/slog"
	"net/http"
	"net/netip"
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/common/httperr"
	v1 "github.com/cubelitblade/community-v2/backend/services/account/api/rest/v1"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/auth"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/authn"
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for authentication.
type Handler struct {
	login    *auth.Login
	logger   *slog.Logger
	validity time.Duration
	policy   CookiePolicy
}

// NewHandler creates and returns a new Handler.
func NewHandler(login *auth.Login, logger *slog.Logger, validity time.Duration, policy CookiePolicy) *Handler {
	logger = logger.With(
		slog.String("service", "account"),
		slog.String("component", "auth/transport/handler"),
	)

	return &Handler{
		login:    login,
		logger:   logger,
		validity: validity,
		policy:   policy,
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
		httperr.WriteBadRequest(c, "Request body must be valid JSON and include username and password.")

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
		h.logger.Debug("login failed",
			slog.String("username", req.Username),
			slog.Any("error", err),
		)

		httperr.WriteMappedError(c, err, authProblem)

		return
	}

	cookie := WriteCookie(v1.AccessTokenCookieName, credentials.Token, h.validity, h.policy)

	http.SetCookie(c.Writer, cookie)
	c.JSON(http.StatusOK, loginResponse{
		ID:   credentials.ID,
		Role: credentials.Role,
	})
}

// Logout handles DELETE /auth/ — erase access token cookie.
func (h *Handler) Logout(c *gin.Context) {
	cookie := WriteCookie(v1.AccessTokenCookieName, "", -1, h.policy)
	http.SetCookie(c.Writer, cookie)
	c.Status(http.StatusNoContent)
}
