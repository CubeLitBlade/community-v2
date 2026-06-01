// Package jwt provides Signer token creation, signing, and parsing.
package jwt

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/golang-jwt/jwt/v5"
)

// SignerOption configures a Signer instance.
type SignerOption func(*Signer)

// Claims represents the Signer claims for an authenticated account.
type Claims struct {
	jwt.RegisteredClaims

	Role string `json:"role"`
}

// Signer create signed JWT access tokens.
type Signer struct {
	key      []byte
	issuer   string
	clock    func() time.Time
	validity time.Duration
	ids      idgen.Generator
}

// NewSigner creates a new Signer instance with the given configuration and ID generator.
func NewSigner(key string, validity time.Duration, ids idgen.Generator, opts ...SignerOption) *Signer {
	signer := &Signer{
		key:      []byte(key),
		issuer:   "",
		clock:    time.Now,
		validity: validity,
		ids:      ids,
	}

	for _, opt := range opts {
		opt(signer)
	}

	return signer
}

// WithSignerClock sets the clock function used for time operations.
func WithSignerClock(clock func() time.Time) SignerOption {
	return func(s *Signer) {
		s.clock = clock
	}
}

// WithSignerIssuer sets the issuer claim for the tokens.
func WithSignerIssuer(issuer string) SignerOption {
	return func(s *Signer) {
		s.issuer = issuer
	}
}

// Sign creates a signed JWT for the given account ID and role.
func (s *Signer) Sign(uid int64, role string) (string, error) {
	id, err := s.ids.NextID()
	if err != nil {
		return "", fmt.Errorf("generate token ID: %w", err)
	}

	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   strconv.FormatInt(uid, 10),
			ExpiresAt: jwt.NewNumericDate(s.clock().Add(s.validity)),
			IssuedAt:  jwt.NewNumericDate(s.clock()),
			ID:        strconv.FormatInt(id, 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(s.key)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

