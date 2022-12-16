package token

import (
	"net/http"
	"strings"

	"github.com/grafana/mimir/pkg/custom/utils"
	"github.com/grafana/mimir/pkg/custom/utils/access"
)

type IPrincipal interface {
	access.ScopeMatcher

	GetClaims() *CustomClaims

	Authenticate()
	IsAuthenticated() bool

	Verify(verifier TokenVerifier) error
	LoadContext(loader AuthContextLoader) error

	WithJWT(jwt string) IPrincipal
	WithTenants(ids ...string) IPrincipal

	WrapRequest(req *http.Request)
}

type principal struct {
	claims  *CustomClaims
	context AuthContext

	jwt           string
	authenticated bool
	tenantIds     []string
}

func NewPrincipal(jwt string, tenants ...string) IPrincipal {
	return &principal{
		jwt:       jwt,
		tenantIds: tenants,
	}
}

func NewAuthorizedPrincipal(tenants ...string) IPrincipal {
	return &principal{
		tenantIds:     tenants,
		authenticated: true,
		context:       NewAuthContext(&trustAllPolicy{}),
	}
}

func (p *principal) GetClaims() *CustomClaims {
	return p.claims
}

func (p *principal) Authenticate() {
	if p.authenticated {
		return
	}

	if p.claims == nil || p.context == nil {
		return
	}

	tenants := p.context.GetPolicy().GetTenants()

	p.tenantIds = utils.MergeTenants(p.tenantIds, tenants)
	p.authenticated = true
}

func (p *principal) IsAuthenticated() bool {
	return p.authenticated
}

func (p *principal) WithJWT(jwt string) IPrincipal {
	p.jwt = jwt
	return p
}

func (p *principal) WithTenants(ids ...string) IPrincipal {
	p.tenantIds = append(p.tenantIds, ids...)
	return p
}

func (p *principal) Verify(verifier TokenVerifier) error {
	verify, err := verifier.Verify(p.jwt)
	if err != nil {
		return err
	}
	p.claims = verify
	return nil
}

func (p *principal) LoadContext(loader AuthContextLoader) error {
	context, err := loader.LoadContext(p.claims)
	if err != nil {
		return err
	}
	p.context = context
	return nil
}

func (p *principal) HasScopes(scopes ...string) bool {
	policy := p.context.GetPolicy()
	if policy == nil {
		return false
	}

	if len(scopes) == 0 {
		return true
	}

	for _, scope := range scopes {
		if !policy.HasScope(scope) {
			return false
		}
	}
	return true
}

func (p *principal) HasAnyScope(scopes ...string) bool {
	policy := p.context.GetPolicy()
	if policy == nil {
		return false
	}

	if len(scopes) == 0 {
		return true
	}

	for _, scope := range scopes {
		if policy.HasScope(scope) {
			return true
		}
	}

	return false
}

func (p *principal) WrapRequest(req *http.Request) {
	req.Header.Set("X-Scope-OrgID", strings.Join(p.tenantIds, "|"))
}
