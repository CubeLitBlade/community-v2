// Package health provides HTTP liveness and readiness endpoints.
package health

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/constant"
)

const (
	statusKey      = "status"
	statusOK       = "ok"
	statusReady    = "ready"
	statusNotReady = "not ready"
)

const pingTimeout = 2 * time.Second

// Pinger checks connectivity to a backend service.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Handler serves health check endpoints.
type Handler struct {
	pinger Pinger
	logger *slog.Logger
}

// NewHandler creates a health check Handler.
func NewHandler(pinger Pinger, logger *slog.Logger) *Handler {
	return &Handler{
		pinger: pinger,
		logger: logger,
	}
}

// RegisterRoutes registers health check routes on the given router.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/healthz", h.liveness)
	router.GET("/readyz", h.readiness)
}

func (h *Handler) liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{statusKey: statusOK})
}

func (h *Handler) readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), pingTimeout)
	defer cancel()

	if err := h.pinger.Ping(ctx); err != nil {
		h.logger.Warn("readiness check failed", slog.Any(constant.LogKeyError, err))
		c.JSON(http.StatusServiceUnavailable, gin.H{statusKey: statusNotReady})

		return
	}

	c.JSON(http.StatusOK, gin.H{statusKey: statusReady})
}
