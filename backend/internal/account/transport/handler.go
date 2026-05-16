// Package transport provides HTTP transport adapters for the account domain.
package transport

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/CubeLitBlade/community-v2/backend/internal/httperr"
)

// AccountRegistrar defines the interface for account-related operations.
type AccountRegistrar interface {
	Register(
		ctx context.Context,
		username string,
		password string,
	) (account.Account, error)
}

// Deps holds the dependencies required for the account Handler.
type Deps struct {
	Registrar AccountRegistrar
	Logger    *slog.Logger
}

// Handler handles HTTP requests for account resources.
type Handler struct {
	registrar AccountRegistrar
	logger    *slog.Logger
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
		registrar: deps.Registrar,
		logger:    deps.Logger,
	}
}

// RegisterRoutes registers the account routes on the given router.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/accounts", h.createAccount)
}

type createAccountRequest struct {
	Username string `binding:"required" json:"username"`
	Password string `binding:"required" json:"password"`
}

func (h *Handler) createAccount(c *gin.Context) {
	var req createAccountRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.WriteInvalidRequest(
			c,
			"Request body must be valid JSON and include"+
				"username and password.",
		)

		return
	}

	acc, err := h.registrar.Register(
		c.Request.Context(),
		req.Username,
		req.Password,
	)
	if err != nil {
		h.logger.Debug(
			"create account failed",
			"username", req.Username,
			"error", err,
		)

		httperr.WriteMappedError(c, err, accountProblem)

		return
	}

	c.Header("Location", fmt.Sprintf("/api/accounts/%d", acc.ID()))
	c.Status(http.StatusCreated)
}
