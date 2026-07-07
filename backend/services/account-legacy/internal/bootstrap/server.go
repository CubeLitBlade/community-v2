package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/pkg/common/httperr"
	"github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/shared"
)

const (
	readHeaderTimeout = time.Second * 5
	readTimeout       = time.Second * 5
	writeTimeout      = time.Second * 5
)

// provideHTTPServer creates an *http.Server, registers lifecycle hooks to
// start listening on app start and gracefully shut down on app stop.
func provideHTTPServer(lifecycle fx.Lifecycle, cfg *shared.Config, router *gin.Engine) *http.Server {
	protection := http.NewCrossOriginProtection()
	protection.SetDenyHandler(http.HandlerFunc(WriteCrossOriginError))
	finalHandler := protection.Handler(router)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           finalHandler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Printf("HTTP server error: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return srv
}

// HTTPServer provides the HTTP server fx module.
func HTTPServer() fx.Option {
	return fx.Options(
		fx.Provide(provideHTTPServer),
	)
}

// WriteCrossOriginError writes a problem+json response for cross-origin request rejections.
func WriteCrossOriginError(w http.ResponseWriter, r *http.Request) {
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
