// Package transport provides HTTP transport adapters for the account domain.
package transport

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/CubeLitBlade/community-v2/backend/internal/authn"
	"github.com/CubeLitBlade/community-v2/backend/internal/httperr"
)

// Registrar defines the interface for account-related operations.
type Registrar interface {
	Register(ctx context.Context, username string, password string) (account.Account, error)
}

// ProfileFinder is the interface for looking up account profiles.
type ProfileFinder interface {
	Find(ctx context.Context, ID int64) (*account.Profile, error)
}

// Deps holds the dependencies required for the account Handler.
type Deps struct {
	Registrar     Registrar
	ProfileFinder ProfileFinder
	Logger        *slog.Logger
}

// Handler handles HTTP requests for account resources.
type Handler struct {
	registrar     Registrar
	profileFinder ProfileFinder
	logger        *slog.Logger
}

// NewHandler creates and returns a new Handler.
func NewHandler(deps Deps) *Handler {
	if deps.Registrar == nil {
		panic("nil registrar")
	}

	if deps.Logger == nil {
		panic("nil logger")
	}

	return &Handler{
		registrar:     deps.Registrar,
		profileFinder: deps.ProfileFinder,
		logger:        deps.Logger,
	}
}

// RegisterRoutes registers the account routes on the given router.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	g := router.Group("/account")

	g.POST("/", h.createAccount)
	g.GET("/", authn.MustAuthenticate(), h.getMyProfile)
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

func (h *Handler) getMyProfile(c *gin.Context) {
	principal, ok := authn.GetPrincipal(c)
	if !ok {
		httperr.WriteInternalServerError(c, httperr.DefaultInternalServerErrorMessage)
	}

	profile, err := h.profileFinder.Find(c.Request.Context(), principal.ID)
	if err != nil {
		if errors.Is(err, account.ErrAccountNotFound) {
			httperr.WriteMappedError(c, err, accountProblem)
		}
	}

	c.JSON(http.StatusOK, profile)
}
