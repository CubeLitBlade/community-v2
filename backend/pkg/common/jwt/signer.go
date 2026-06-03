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

// Signer create signed JWT access tokens.
type Signer struct {
	key    []byte
	issuer string
	clock  func() time.Time
	ids    idgen.Generator
}

// NewSigner creates a new Signer instance with the given configuration and ID generator.
func NewSigner(key []byte, ids idgen.Generator, opts ...SignerOption) *Signer {
	signer := &Signer{
		key:    key,
		issuer: "",
		clock:  time.Now,
		ids:    ids,
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

// NewClaims creates a new JWT RegisteredClaims with the given subject and TTL.
func (s *Signer) NewClaims(subject string, ttl time.Duration) (jwt.RegisteredClaims, error) {
	id, err := s.ids.NextID()
	if err != nil {
		return jwt.RegisteredClaims{}, fmt.Errorf("generate token ID: %w", err)
	}

	now := s.clock()

	return jwt.RegisteredClaims{
		Issuer:    s.issuer,
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        strconv.FormatInt(id, 10),
	}, nil
}

// Sign creates a signed JWT for the given claims.
func (s *Signer) Sign(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(s.key)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}
