package bootstrap

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
)

func newHTTPServer(cfg Config, router *gin.Engine) (*http.Server, error) {
	handler, err := newHTTPHandler(cfg, router)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:    cfg.Addr,
		Handler: handler,
	}, nil
}

func newHTTPHandler(cfg Config, router *gin.Engine) (http.Handler, error) {
	if len(cfg.CSRFAuthKey) != csrfAuthKeyBytes {
		return nil, errors.New("CSRF_AUTH_KEY must be exactly 32 bytes")
	}

	csrfMiddleware := csrf.Protect(
		[]byte(cfg.CSRFAuthKey),
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

	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":   "about:blank",
		"title":  "CSRF token invalid",
		"status": http.StatusForbidden,
		"detail": "The CSRF token is missing or invalid.",
	})
}
