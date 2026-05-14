package bootstrap

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func newRouter() (*gin.Engine, error) {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	err := r.SetTrustedProxies(nil)
	if err != nil {
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}

	return r, nil
}
