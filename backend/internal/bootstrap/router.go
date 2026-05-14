package bootstrap

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func newRouter() (*gin.Engine, error) {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}

	return r, nil
}
