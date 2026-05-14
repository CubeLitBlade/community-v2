package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/gin-gonic/gin"
)

type AccountService interface {
	CreateAccount(ctx context.Context, username string, password string) (account.Account, error)
}

type AccountHandler struct {
	accounts AccountService
	logger   *slog.Logger
}

func NewAccountHandler(accounts AccountService, logger *slog.Logger) *AccountHandler {
	if accounts == nil {
		panic("web.NewAccountHandler: accounts is nil")
	}

	if logger == nil {
		panic("web.NewAccountHandler: logger is nil")
	}

	return &AccountHandler{
		accounts: accounts,
		logger:   logger,
	}
}

func (h *AccountHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/accounts", h.createAccount)
}

type createAccountRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AccountHandler) createAccount(c *gin.Context) {
	var req createAccountRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		writeInvalidRequest(c, "Request body must be valid JSON and include username and password.")
		return
	}

	acc, err := h.accounts.CreateAccount(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.logger.Warn(
			"create account failed",
			"username", req.Username,
			"error", err,
		)

		writeMappedError(c, err, accountProblem)
		return
	}

	c.Header("Location", fmt.Sprintf("/api/accounts/%d", acc.ID()))
	c.Status(http.StatusCreated)
}
