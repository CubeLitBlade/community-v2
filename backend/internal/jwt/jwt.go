package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"

	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
)

// Claims represents the JWT claims for an authenticated user.
type Claims struct {
	gojwt.RegisteredClaims

	Role string `json:"role"`
}

// Token represents a signed JWT access token.
type Token string

// Config holds the configuration parameters for JWT issuance.
type Config struct {
	Key      string
	Issuer   string
	Validity time.Duration
}

// Issuer creates signed JWT access tokens.
type Issuer struct {
	key      []byte
	issuer   string
	now      func() time.Time
	validity time.Duration
	ids      idgen.Generator
}

// NewIssuer creates a new Issuer with the given configuration and ID generator.
func NewIssuer(
	cfg *Config, ids idgen.Generator,
) *Issuer {
	return &Issuer{
		key:      []byte(cfg.Key),
		issuer:   cfg.Issuer,
		now:      time.Now,
		validity: cfg.Validity,
		ids:      ids,
	}
}

// Issue creates a signed JWT for the given user ID and role.
func (i *Issuer) Issue(uid int64, role string) (string, error) {
	id, err := i.ids.NextID()
	if err != nil {
		return "", fmt.Errorf(
			"could not generate token ID: %w",
			err,
		)
	}

	claims := Claims{
		Role: role,
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   strconv.FormatInt(uid, 10),
			ExpiresAt: gojwt.NewNumericDate(i.now().Add(i.validity)),
			IssuedAt:  gojwt.NewNumericDate(i.now()),
			ID:        strconv.FormatInt(id, 10),
		},
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(i.key)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

var errTokenNotValid = errors.New("token is not valid")

// Parse validates and parses a JWT token string, returning its claims.
func Parse(tokenString string, key []byte) (*Claims, error) {
	claims := &Claims{
		RegisteredClaims: gojwt.RegisteredClaims{},
		Role:             "",
	}

	parsed, err := gojwt.ParseWithClaims(
		tokenString, claims,
		func(_ *gojwt.Token) (any, error) {
			return key, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if !parsed.Valid {
		return nil, errTokenNotValid
	}

	return claims, nil
}
