package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"

	"github.com/CubeLitBlade/community-v2/backend/shared/idgen"
)

var (
	// ErrInvalidToken is returned when the token is invalid.
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenExpired is returned when the token is expired.
	ErrTokenExpired = errors.New("token is expired")
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

// JWT creates and parses signed JWT access tokens.
type JWT struct {
	key      []byte
	issuer   string
	now      func() time.Time
	validity time.Duration
	ids      idgen.Generator
}

// New creates a new JWT with the given configuration and ID generator.
func New(
	cfg *Config, ids idgen.Generator,
) *JWT {
	return &JWT{
		key:      []byte(cfg.Key),
		issuer:   cfg.Issuer,
		now:      time.Now,
		validity: cfg.Validity,
		ids:      ids,
	}
}

// Issue creates a signed JWT for the given user ID and role.
func (j *JWT) Issue(uid int64, role string) (string, error) {
	id, err := j.ids.NextID()
	if err != nil {
		return "", fmt.Errorf(
			"could not generate token ID: %w",
			err,
		)
	}

	claims := Claims{
		Role: role,
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   strconv.FormatInt(uid, 10),
			ExpiresAt: gojwt.NewNumericDate(j.now().Add(j.validity)),
			IssuedAt:  gojwt.NewNumericDate(j.now()),
			ID:        strconv.FormatInt(id, 10),
		},
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(j.key)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

// Parse validates and parses a JWT token string, returning its claims.
func Parse(tokenString string, key string) (*Claims, error) {
	claims := &Claims{
		RegisteredClaims: gojwt.RegisteredClaims{},
		Role:             "",
	}

	token, err := gojwt.ParseWithClaims(
		tokenString, claims,
		func(token *gojwt.Token) (any, error) {
			return []byte(key), nil
		},
	)
	if err != nil {
		if errors.Is(err, gojwt.ErrTokenExpired) {
			return nil, ErrInvalidToken
		}

		return nil, ErrInvalidToken
	} else if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
