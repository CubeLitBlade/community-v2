package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
)

// Claims represents the JWT claims for an authenticated user.
type Claims struct {
	jwt.RegisteredClaims

	Role string `json:"role"`
}

// Token represents a signed JWT access token.
type Token string

// JWTConfig holds the configuration parameters for JWT issuance.
type JWTConfig struct {
	Key      string
	Issuer   string
	Validity time.Duration
}

// JWTIssuer creates signed JWT access tokens.
type JWTIssuer struct {
	key      []byte
	issuer   string
	now      func() time.Time
	validity time.Duration
	ids      idgen.Generator
}

// NewJWTIssuer creates a new JWTIssuer with the given configuration and ID generator.
func NewJWTIssuer(
	jwtConfig *JWTConfig, ids idgen.Generator,
) *JWTIssuer {
	return &JWTIssuer{
		key:      []byte(jwtConfig.Key),
		issuer:   jwtConfig.Issuer,
		now:      time.Now,
		validity: jwtConfig.Validity,
		ids:      ids,
	}
}

// Issue creates a signed JWT for the given user ID and role.
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
			ID:        strconv.FormatInt(id, 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(i.key)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}
