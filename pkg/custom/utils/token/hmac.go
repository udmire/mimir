package token

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

type HMACTokenSignerVerifier interface {
	SetSecret(secret []byte)
}

type hmacTokenHandler struct {
	secret []byte
}

func (h *hmacTokenHandler) SetSecret(secret []byte) {
	h.secret = secret
}

type hmacTokenSigner struct {
	hmacTokenHandler
}

func (h *hmacTokenSigner) Sign(claims *CustomClaims) (string, error) {
	if claims == nil {
		return "", fmt.Errorf("claims need to provided")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.secret)
}

type hmacTokenVerifier struct {
	hmacTokenHandler
}

func (h *hmacTokenVerifier) Verify(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return h.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	if c, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return c, nil
	} else {
		return nil, errors.New("invalid token")
	}
}

func NewSigner(secret []byte) TokenSigner {
	return &hmacTokenSigner{
		hmacTokenHandler: hmacTokenHandler{
			secret: secret,
		},
	}
}

func NewVerifier(secret []byte) TokenVerifier {
	return &hmacTokenVerifier{
		hmacTokenHandler{
			secret: secret,
		},
	}
}
