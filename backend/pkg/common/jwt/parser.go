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

type ParserOption func(*Parser)

type Parser struct {
	key    []byte
	issuer string
	clock  func() time.Time
}

func NewParser(key string, opts ...ParserOption) *Parser {
	p := &Parser{
		key:    []byte(key),
		issuer: "",
		clock:  time.Now,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func WithParserIssuer(issuer string) ParserOption {
	return func(p *Parser) {
		p.issuer = issuer
	}
}

func WithParserClock(clock func() time.Time) ParserOption {
	return func(p *Parser) {
		p.clock = clock
	}
}

func (p *Parser) Parse(value string) (*Claims, error) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		Role:             "",
	}

	token, err := jwt.ParseWithClaims(value, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}

		return p.key, nil
	}, jwt.WithTimeFunc(p.clock),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}

		return nil, ErrInvalidToken
	} else if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
