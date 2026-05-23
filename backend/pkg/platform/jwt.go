package platform

import (
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/common/jwt"
)

// JWTConfig supplies the secret key and access token TTL for JWT issuance.
type JWTConfig interface {
	JWTSecret() string
	AccessTokenTTL() time.Duration
}

// ProvideJWTConfig creates a *jwt.Config from the given JWTConfig interface.
func ProvideJWTConfig(cfg JWTConfig) *jwt.Config {
	return &jwt.Config{
		Key:      cfg.JWTSecret(),
		Issuer:   "community-v2",
		Validity: cfg.AccessTokenTTL(),
	}
}
