// Package server provides Gin engine setup, HTTP server lifecycle management,
// and request-level middleware.
package server

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/platform/config"
	"github.com/gin-gonic/gin"
)

// NewGinEngine creates a gin.Engine with common middleware applied.
func NewGinEngine(cfg *config.GinConfig, logger *slog.Logger) *gin.Engine {
	gin.SetMode(cfg.GinMode)

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(requestIDMiddleware())

	if cfg.GinMode == gin.DebugMode {
		router.Use(gin.Logger())
	} else {
		router.Use(loggerMiddleware(logger))
	}

	// Configure trusted proxies
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		// This shouldn't happen with a valid config; panic to fail fast
		panic(fmt.Sprintf("set trusted proxies: %v", err))
	}

	return router
}

// requestIDMiddleware assigns a unique request ID to each request.
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Request-ID") == "" {
			c.Request.Header.Set("X-Request-ID", strconv.FormatInt(time.Now().UnixNano(), 10))
		}
		c.Header("X-Request-ID", c.GetHeader("X-Request-ID"))
		c.Next()
	}
}

// loggerMiddleware logs each request using the structured logger.
func loggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method

		attrs := []slog.Attr{
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("query", query),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
		}

		//nolint:mnd // 500 is the HTTP status threshold for server errors.
		if status >= 500 {
			logger.LogAttrs(c.Request.Context(), slog.LevelError, "request completed", attrs...)
		} else {
			logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "request completed", attrs...)
		}
	}
}
