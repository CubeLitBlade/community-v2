package bootstrap

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func newRouter() (*gin.Engine, error) {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	err := router.SetTrustedProxies(nil)
	if err != nil {
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}

	return router, nil
}
