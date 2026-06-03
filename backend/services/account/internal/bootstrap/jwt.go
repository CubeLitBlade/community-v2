package bootstrap

import (
	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/pkg/common/jwt"
	"go.uber.org/fx"
)

// MinJWTKeyLength is the minimum required length in bytes for a JWT signing key.
const MinJWTKeyLength = 32

// IssuerProvider provides the token issuer string.
type IssuerProvider interface {
	Issuer() string
}

// JWTConfig holds the configuration for JWT signing and parsing.
type JWTConfig struct {
	Key string `env:"JWT_SECRET" required:"true"`
}

func provideJWT(cfg JWTConfig, issuer IssuerProvider, ids idgen.Generator) (*jwt.Signer, *jwt.Parser) {
	key := []byte(cfg.Key)
	if len(key) < MinJWTKeyLength {
		panic("JWT key must be at least 32 bytes long")
	}

	signer := jwt.NewSigner(key, ids, jwt.WithSignerIssuer(issuer.Issuer()))
	parser := jwt.NewParser(key, jwt.WithParserIssuer(issuer.Issuer()))

	return signer, parser
}

// JWTModule provides the JWT signer and parser fx module.
func JWTModule() fx.Option {
	return fx.Options(
		fx.Provide(provideJWT),
	)
}
