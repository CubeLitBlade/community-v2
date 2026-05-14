package bootstrap

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/CubeLitBlade/community-v2/backend/internal/web"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
)

const httpReadHeaderTimeout = 5 * time.Second

func newHTTPServer(cfg Config, router *gin.Engine) (*http.Server, error) {
	handler, err := newHTTPHandler(cfg, router)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: httpReadHeaderTimeout,
	}, nil
}

func newHTTPHandler(cfg Config, router *gin.Engine) (http.Handler, error) {
	if len(cfg.CSRFAuthKey) != csrfAuthKeyBytes {
		return nil, errCSRFAuthKeyLength
	}

	csrfMiddleware := csrf.Protect(
		cfg.CSRFAuthKey,
		csrf.CookieName("csrf_token"),
		csrf.RequestHeader("X-CSRF-Token"),
		csrf.Path("/"),
		csrf.Secure(cfg.CookieSecure),
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.ErrorHandler(http.HandlerFunc(writeCSRFError)),
	)

	return csrfMiddleware(router), nil
}

func writeCSRFError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusForbidden)

	response := web.ProblemDetail{
		Type:     "about:blank",
		Title:    "CSRF token invalid",
		Status:   http.StatusForbidden,
		Detail:   "The CSRF token is missing or invalid.",
		Instance: r.URL.Path,
		Code:     "CSRF_TOKEN_INVALID",
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}
