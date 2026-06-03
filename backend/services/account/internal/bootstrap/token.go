package bootstrap

import (
	"time"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/contracts"
)

var _ contracts.TTLProvider = (*TokenConfig)(nil)

// TokenConfig holds the token issuer and TTL configuration.
type TokenConfig struct {
	IssuerName   string        `env:"JWT_ISSUER" envDefault:"community-v2"`
	AccessExpiry time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"2h"`
}

// AccessTokenTTL returns the access token time-to-live.
func (cfg TokenConfig) AccessTokenTTL() time.Duration {
	return cfg.AccessExpiry
}

// Issuer returns the token issuer string.
func (cfg TokenConfig) Issuer() string {
	return cfg.IssuerName
}
