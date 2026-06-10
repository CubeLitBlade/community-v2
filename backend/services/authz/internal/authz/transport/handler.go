package transport

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cubelitblade/community-v2/backend/pkg/common/httperr"
	v1 "github.com/cubelitblade/community-v2/backend/services/authz/api/types/v1"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/constant"
)

// Checker verifies if a user has a specific relation to an object.
type Checker interface {
	Check(ctx context.Context, user, relation, object string) (bool, error)
}

// Handler handles HTTP requests for the authz service.
type Handler struct {
	checker Checker
	logger  *slog.Logger
}

// NewHandler creates a new Handler with the given Checker and logger.
func NewHandler(checker Checker, logger *slog.Logger) *Handler {
	logger = logger.With(
		slog.String(constant.LogKeyService, constant.LogServiceAuthz),
		slog.String(constant.LogKeyComponent, constant.LogComponentTransportHandler),
	)

	return &Handler{
		checker: checker,
		logger:  logger,
	}
}

// RegisterRoutes registers the authz HTTP routes on the given router.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	g := router.Group("/authz")

	g.POST("/check", h.check)
}

func (h *Handler) check(c *gin.Context) {
	var req v1.AuthorizationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.WriteBadRequest(c, detailInvalidCheckBody())

		return
	}

	ok, err := h.checker.Check(c.Request.Context(), req.User, req.Relation, req.Object)
	if err != nil {
		h.logger.Error("failed to authorize",
			slog.String("user", req.User),
			slog.String("relation", req.Relation),
			slog.String("object", req.Object),
			slog.String("client_ip", c.ClientIP()),
			slog.String("path", c.Request.URL.Path),
			slog.String("trace_id", c.GetString("X-Trace-ID")),
		)
		httperr.WriteMappedError(c, err, authzProblem)

		return
	}

	c.JSON(http.StatusOK, v1.Decision{
		Allowed: ok,
	})
}
