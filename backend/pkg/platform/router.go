package platform

import "github.com/gin-gonic/gin"

// HTTPMounter is implemented by modules that mount HTTP routes.
type HTTPMounter interface {
	RegisterRoutes(router gin.IRouter)
}

// RegisterRouters invokes RegisterRoutes on each mounter using the given router.
func RegisterRouters(router gin.IRouter, mounters []HTTPMounter) {
	for _, m := range mounters {
		m.RegisterRoutes(router)
	}
}
