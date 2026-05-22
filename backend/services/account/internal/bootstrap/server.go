package bootstrap

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/common/httperr"
	"github.com/gin-gonic/gin"
)

const httpReadHeaderTimeout = 5 * time.Second

func newHTTPServer(cfg *Config, router *gin.Engine) *http.Server {
	handler := newHTTPHandler(router)

	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: httpReadHeaderTimeout,
	}
}

func newHTTPHandler(router *gin.Engine) http.Handler {
	protection := http.NewCrossOriginProtection()
	protection.SetDenyHandler(http.HandlerFunc(writeCrossOriginError))

	return protection.Handler(router)
}

func writeCrossOriginError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusForbidden)

	err := json.NewEncoder(w).Encode(httperr.ProblemDetail{
		Type:     "about:blank",
		Title:    "Cross-origin request rejected",
		Status:   http.StatusForbidden,
		Detail:   "The request was blocked by cross-origin protection.",
		Instance: r.URL.Path,
		Code:     "CROSS_ORIGIN_REJECTED",
	})
	if err != nil {
		log.Printf(
			"Failed to write cross-origin error response: %v", err,
		)
	}
}
