package bootstrap

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/auth/transport"
	"github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/shared"
)

var errUnsupportedSamesite = errors.New("unsupported samesite")

// CookiePolicy implements transport.CookiePolicy based on the runtime environment.
type CookiePolicy struct {
	secure   bool
	sameSite http.SameSite
}

// Secure returns whether cookies should be marked as Secure.
func (p CookiePolicy) Secure() bool {
	return p.secure
}

// SameSite returns the SameSite attribute for cookies.
func (p CookiePolicy) SameSite() http.SameSite {
	return p.sameSite
}

func provideCookiePolicy(cfg *shared.CookieConfig) (*CookiePolicy, error) {
	var sameSite http.SameSite

	switch strings.ToLower(cfg.SameSite) {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "none":
		sameSite = http.SameSiteNoneMode
	default:
		return nil, fmt.Errorf("%w: %q", errUnsupportedSamesite, cfg.SameSite)
	}

	return &CookiePolicy{
		secure:   cfg.Secure,
		sameSite: sameSite,
	}, nil
}

// CookieModule provides the cookie policy fx module.
func CookieModule() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(provideCookiePolicy, fx.As(new(transport.CookiePolicy))),
		),
	)
}
