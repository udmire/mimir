package auth

import (
	"github.com/grafana/mimir/pkg/cache"
	"github.com/grafana/mimir/pkg/custom/utils/token"
)

type PrincipalVerifierWrapper struct {
	cache    cache.Cache
	delegate token.TokenVerifier
}

func (p *PrincipalVerifierWrapper) Verify(tokenString string) (*token.CustomClaims, error) {
	// TODO read from the cache first
	return p.delegate.Verify(tokenString)
}

// func (p *PrincipalVerifierWrapper) Verify(principal token.IPrincipal) error {
// 	// TODO read from cache first
// 	return p.delegate.Verify(principal)
// }
