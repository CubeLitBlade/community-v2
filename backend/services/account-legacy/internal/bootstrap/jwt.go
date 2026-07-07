package bootstrap

import (
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/pkg/common/jwt"
	"github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/shared"
)

// MinJWTKeyLength is the minimum required length in bytes for a JWT signing key.
const MinJWTKeyLength = 32

func provideJWT(jwtCfg *shared.JWTConfig, tokenCfg *shared.TokenConfig, ids idgen.Generator) (*jwt.Signer, *jwt.Parser) {
	key := []byte(jwtCfg.Secret)
	if len(key) < MinJWTKeyLength {
		panic("JWT key must be at least 32 bytes long")
	}

	signer := jwt.NewSigner(key, ids, jwt.WithSignerIssuer(tokenCfg.TokenIssuer))
	parser := jwt.NewParser(key, jwt.WithParserIssuer(tokenCfg.TokenIssuer))

	return signer, parser
}

// JWTModule provides the JWT signer and parser fx module.
func JWTModule() fx.Option {
	return fx.Options(
		fx.Provide(provideJWT),
	)
}
