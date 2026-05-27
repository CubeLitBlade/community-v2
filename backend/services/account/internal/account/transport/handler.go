// Package transport provides HTTP transport adapters for the account domain.
package transport

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/cubelitblade/community-v2/backend/pkg/common/httperr"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/account"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/authn"
	"github.com/gin-gonic/gin"
)

// Registrar defines the interface for account-related operations.
type Registrar interface {
	Register(ctx context.Context, username string, password string) (account.Account, error)
}

// ProfileFinder is the interface for looking up account profiles.
type ProfileFinder interface {
	Find(ctx context.Context, ID int64) (*account.Profile, error)
}

// Handler handles HTTP requests for account resources.
type Handler struct {
	registrar     Registrar
	profileFinder ProfileFinder
	logger        *slog.Logger
}

// NewHandler creates and returns a new Handler.
func NewHandler(
	registrar Registrar,
	finder ProfileFinder,
	logger *slog.Logger,
) *Handler {
	if logger == nil {
		logger = slog.Default()
		logger.Warn("No logger provided, using default logger")
	}

	return &Handler{
		registrar:     registrar,
		profileFinder: finder,
		logger:        logger,
	}
}

// RegisterRoutes registers the account routes on the given router.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	g := router.Group("/accounts")

	g.POST("/", h.createAccount)
	g.GET("/", authn.MustAuthenticate(), h.getOwnProfile)
	g.GET("/:id", h.getProfile)
}

type createAccountRequest struct {
	Username string `binding:"required" json:"username"`
	Password string `binding:"required" json:"password"`
}

func (h *Handler) createAccount(c *gin.Context) {
	var req createAccountRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.WriteBadRequest(c, detailInvalidCreateAccountBody())

		return
	}

	acc, err := h.registrar.Register(
		c.Request.Context(),
		req.Username,
		req.Password,
	)
	if err != nil {
		h.logger.Debug("create account failed", "username", req.Username, "error", err)
		httperr.WriteMappedError(c, err, accountProblem)

		return
	}

	c.Header("Location", fmt.Sprintf("/api/accounts/%d", acc.ID()))
	c.Status(http.StatusCreated)
}

func (h *Handler) getOwnProfile(c *gin.Context) {
	principal, ok := authn.GetPrincipal(c)
	if !ok {
		httperr.WriteInternalServerError(c, httperr.DefaultInternalServerErrorMessage)

		return
	}

	profile, err := h.profileFinder.Find(c.Request.Context(), principal.ID)
	if err != nil {
		if errors.Is(err, account.ErrAccountNotFound) {
			httperr.WriteMappedError(c, err, accountProblem)

			return
		}

		httperr.WriteInternalServerError(c, httperr.DefaultInternalServerErrorMessage)

		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *Handler) getProfile(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httperr.WriteBadRequest(c, httperr.DefaultBadRequestMessage)

		return
	}

	profile, err := h.profileFinder.Find(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, account.ErrAccountNotFound) {
			httperr.WriteMappedError(c, err, accountProblem)

			return
		}

		httperr.WriteInternalServerError(c, httperr.DefaultInternalServerErrorMessage)

		return
	}

	c.JSON(http.StatusOK, profile)
}
