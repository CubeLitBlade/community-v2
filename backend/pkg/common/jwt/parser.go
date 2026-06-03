// Package jwt provides Signer token creation, signing, and parsing.
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Sentinel errors for jwt.
var (
	ErrInvalidToken            = errors.New("invalid token")
	ErrTokenExpired            = errors.New("token is expired")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
)

// ParserOption is a functional option for configuring a Parser.
type ParserOption func(*Parser)

// Parser validates and decodes JWT tokens.
type Parser struct {
	key    []byte
	issuer string
	clock  func() time.Time
}

// NewParser creates a new Parser with the given key and optional configuration.
func NewParser(key []byte, opts ...ParserOption) *Parser {
	parser := &Parser{
		key:    key,
		issuer: "",
		clock:  time.Now,
	}

	for _, opt := range opts {
		opt(parser)
	}

	return parser
}

// WithParserIssuer sets the expected issuer for the parser.
func WithParserIssuer(issuer string) ParserOption {
	return func(p *Parser) {
		p.issuer = issuer
	}
}

// WithParserClock sets the clock function used for time validation.
func WithParserClock(clock func() time.Time) ParserOption {
	return func(p *Parser) {
		p.clock = clock
	}
}

// Parse validates the given JWT string and populates the provided claims.
func (p *Parser) Parse(value string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(value, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}

		return p.key, nil
	}, jwt.WithTimeFunc(p.clock), jwt.WithIssuer(p.issuer))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return ErrTokenExpired
		}

		return ErrInvalidToken
	}

	if !token.Valid {
		return ErrInvalidToken
	}

	return nil
}
