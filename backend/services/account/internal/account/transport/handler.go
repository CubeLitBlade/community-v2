// Package transport provides HTTP transport adapters for the account domain.
package transport

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/cubelitblade/community-v2/backend/pkg/common/httperr"
	v1 "github.com/cubelitblade/community-v2/backend/services/account/api/rest/v1"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/account"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/authn"
)

// Registrar defines the interface for account-related operations.
type Registrar interface {
	Register(ctx context.Context, username string, password string) (int64, error)
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
func NewHandler(registrar Registrar, finder ProfileFinder, logger *slog.Logger) *Handler {
	logger = logger.With(
		slog.String("service", "account"),
		slog.String("component", "transport/handler"),
	)

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

func (h *Handler) createAccount(c *gin.Context) {
	var req v1.CreateAccountRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.WriteBadRequest(c, detailInvalidCreateAccountBody())

		return
	}

	id, err := h.registrar.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.logger.Debug("create account failed",
			slog.String("username", req.Username),
			slog.Any("error", err),
		)
		httperr.WriteMappedError(c, err, accountProblem)

		return
	}

	c.Header("Location", fmt.Sprintf("%s/%d", c.Request.URL.Path, id))
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

	c.JSON(http.StatusOK, v1.Profile{
		ID:          profile.ID,
		Username:    profile.Username,
		DisplayName: profile.DisplayName,
	})
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

	c.JSON(http.StatusOK, v1.Profile{
		ID:          profile.ID,
		Username:    profile.Username,
		DisplayName: profile.DisplayName,
	})
}
