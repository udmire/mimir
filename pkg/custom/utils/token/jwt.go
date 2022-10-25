package token

import (
	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
}

func (c *CustomClaims) GetPolicy() string {
	return c.Subject
}

type TokenSigner interface {
	Sign(claim *CustomClaims) (string, error)
}

type TokenVerifier interface {
	Verify(tokenString string) (*CustomClaims, error)
}

type TokenSignerVerifier interface {
	TokenSigner
	TokenVerifier
}
