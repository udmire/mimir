package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-kit/log"
	"github.com/grafana/mimir/pkg/custom/utils/token"
	"github.com/prometheus/prometheus/model/relabel"
)

type oidcPrincipalVerifier struct {
	logger log.Logger

	cfg          OidcConfig
	regexp       *relabel.Regexp
	oidcVerifier *oidc.IDTokenVerifier
}

func NewOidcPrincipalVerifier(cfg OidcConfig, logger log.Logger) (token.TokenVerifier, error) {
	provider, err := oidc.NewProvider(context.Background(), cfg.IssuerUrl)
	if err != nil {
		return nil, err
	}

	oidcCfg := &oidc.Config{}
	if cfg.Client.ClientID == "" {
		oidcCfg.SkipClientIDCheck = true
	} else {
		oidcCfg.ClientID = cfg.Client.ClientID
	}

	regexp, err := relabel.NewRegexp(cfg.AccessPolicyRegex)
	if err != nil {
		return nil, err
	}

	o := &oidcPrincipalVerifier{
		cfg:          cfg,
		oidcVerifier: provider.Verifier(oidcCfg),
		regexp:       &regexp,
	}
	o.logger = logger
	return o, nil
}

func (o *oidcPrincipalVerifier) Verify(tokenString string) (*token.CustomClaims, error) {
	idToken, err := o.oidcVerifier.Verify(context.Background(), tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to verify JWT token: %w", err)
	}

	claims := &token.CustomClaims{}
	if err = idToken.Claims(&claims); err != nil {
		return nil, err
	}

	return claims, nil
}