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

	cfg          *OidcConfig
	regexp       *relabel.Regexp
	oidcVerifier *oidc.IDTokenVerifier
}

func NewOidcPrincipalVerifier(cfg *OidcConfig, logger log.Logger) (token.TokenVerifier, error) {
	provider, err := oidc.NewProvider(context.Background(), cfg.IssuerUrl)
	if err != nil {
		return nil, err
	}

	oidcCfg := &oidc.Config{}
	if cfg.Client == nil || cfg.Client.ClientID == "" {
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

// func (o *oidcPrincipalVerifier) Verify(principal *Principal) error {
// 	token, err := o.oidcVerifier.Verify(context.Background(), principal.JWT)
// 	if err != nil {
// 		return fmt.Errorf("failed to verify JWT token: %w", err)
// 	}
//
// 	// Validate with OP
// 	claims := new(json.RawMessage)
// 	if err = token.Claims(&claims); err != nil {
// 		return err
// 	}
// 	jsonNode, err := simplejson.NewJson(*claims)
// 	if err != nil {
// 		return err
// 	}
//
// 	var policies []string
// 	jsonNode, exists := jsonNode.CheckGet(o.cfg.AccessPolicyClaim)
// 	if !exists {
// 		if o.cfg.DefaultAccessPolicy != "" {
// 			policies = append(policies, o.cfg.DefaultAccessPolicy)
// 		}
// 	} else {
// 		array, err := jsonNode.StringArray()
// 		if err != nil {
// 			s, err := jsonNode.String()
// 			if err != nil {
// 				level.Warn(o.logger).Log("msg", "invalid configuration", "err", err)
// 				return err
// 			}
// 			array = append(array, s)
// 		}
//
// 		if o.regexp != nil {
// 			for _, claim := range array {
// 				submatch := o.regexp.FindStringSubmatch(claim)
// 				policies = append(policies, submatch[0])
// 			}
// 		}
// 	}
//
// 	return o.Process(policies, principal)
// }
