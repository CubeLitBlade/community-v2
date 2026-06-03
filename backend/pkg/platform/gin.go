package platform

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// NewGinEngine creates a gin.Engine with contracts middleware applied.
func NewGinEngine(handlers ...gin.HandlerFunc) *gin.Engine {
	engine := gin.New()
	engine.Use(handlers...)

	return engine
}

// RequestIDMiddleware assigns a unique request ID to each request.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Request-ID") == "" {
			c.Request.Header.Set("X-Request-ID", strconv.FormatInt(time.Now().UnixNano(), 10))
		}

		c.Header("X-Request-ID", c.GetHeader("X-Request-ID"))
		c.Next()
	}
}
