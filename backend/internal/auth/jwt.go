package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
)

// Claims represents the JWT claims for an authenticated user.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type IDGenerator interface {
	NextID() (int64, error)
}

type JWTConfig struct {
	Key      string
	Issuer   string
	Validity time.Duration
}

type JWTIssuer struct {
	key      []byte
	issuer   string
	now      func() time.Time
	validity time.Duration
	ids      IDGenerator
}

func NewJWTIssuer(
	jwtConfig *JWTConfig, ids IDGenerator,
) *JWTIssuer {
	return &JWTIssuer{
		key:      []byte(jwtConfig.Key),
		issuer:   jwtConfig.Issuer,
		now:      time.Now,
		validity: jwtConfig.Validity,
		ids:      ids,
	}
}

func (i *JWTIssuer) Issue(uid account.ID, role account.Role) (string, error) {
	id, err := i.ids.NextID()
	if err != nil {
		return "", fmt.Errorf(
			"could not generate token ID: %w",
			err,
		)
	}

	claims := Claims{
		Role: role.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   fmt.Sprintf("%d", uid),
			Audience:  nil,
			ExpiresAt: jwt.NewNumericDate(i.now().Add(i.validity)),
			NotBefore: nil,
			IssuedAt:  jwt.NewNumericDate(i.now()),
			ID:        fmt.Sprintf("%d", id),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(i.key)
}
