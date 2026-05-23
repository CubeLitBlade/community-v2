// Package health provides HTTP liveness and readiness endpoints.
package health

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	statusKey      = "status"
	statusOK       = "ok"
	statusReady    = "ready"
	statusNotReady = "not ready"
)

// Handler serves health check endpoints.
type Handler struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewHandler creates a health check Handler.
func NewHandler(db *gorm.DB, logger *slog.Logger) *Handler {
	return &Handler{
		db:     db,
		logger: logger,
	}
}

// RegisterRoutes registers health check routes on the given router.
// These are mounted directly on the engine, outside of any auth middleware.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/healthz", h.liveness)
	router.GET("/readyz", h.readiness)
}

func (h *Handler) liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{statusKey: statusOK})
}

func (h *Handler) readiness(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		h.logger.Error("failed to get database handle for readiness check", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{statusKey: statusNotReady})

		return
	}

	if err := sqlDB.PingContext(c.Request.Context()); err != nil {
		h.logger.Warn("database ping failed for readiness check", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{statusKey: statusNotReady})

		return
	}

	c.JSON(http.StatusOK, gin.H{statusKey: statusReady})
}
