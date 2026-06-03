package bootstrap

import (
	"net/http"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/auth/transport"
	"go.uber.org/fx"
)

// CookiePolicy implements transport.CookiePolicy based on the runtime environment.
type CookiePolicy struct {
	secure   bool
	sameSite http.SameSite
}

// CookieTTLConfig holds the TTL configuration for cookies.
type CookieTTLConfig struct {
	Access time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"2h"`
}

// Secure returns whether cookies should be marked as Secure.
func (p CookiePolicy) Secure() bool {
	return p.secure
}

// SameSite returns the SameSite attribute for cookies.
func (p CookiePolicy) SameSite() http.SameSite {
	return p.sameSite
}

func provideCookiePolicy(appRT AppRuntime) CookiePolicy {
	switch appRT.Environment() {
	case AppEnvDevelopment:
		return CookiePolicy{
			secure:   false,
			sameSite: http.SameSiteLaxMode,
		}
	case AppEnvProduction:
		return CookiePolicy{
			secure:   true,
			sameSite: http.SameSiteLaxMode,
		}
	}

	return CookiePolicy{
		secure:   false,
		sameSite: http.SameSiteLaxMode,
	}
}

// CookieModule provides the cookie policy fx module.
func CookieModule() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(provideCookiePolicy, fx.As(new(transport.CookiePolicy))),
		),
	)
}
