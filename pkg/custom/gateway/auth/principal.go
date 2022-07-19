package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/mimir/pkg/custom/utils/token"
)

type principalCtxKey struct {
}

// GetPrincipal gets the principal from the context.
func GetPrincipal(ctx context.Context) token.IPrincipal {
	principal, ok := ctx.Value(principalCtxKey{}).(token.IPrincipal)
	if ok {
		return principal
	}

	return nil
}

// SetPrincipal sets the principal into the context.
func SetPrincipal(ctx context.Context, p token.IPrincipal) context.Context {
	return context.WithValue(ctx, principalCtxKey{}, p)
}

type PrincipalReader interface {
	// Principal extracts a principal from the http.Request.
	// It's not an error for there to be no principal in the request.
	Principal(r *http.Request) (token.IPrincipal, error)

	CanProcess(r *http.Request) bool
}

// Basic Auth Principal Reader
type basicAuthPrincipalReader struct {
	log      log.Logger
	verifier token.TokenVerifier
}

func (b *basicAuthPrincipalReader) Principal(req *http.Request) (token.IPrincipal, error) {
	auth, password, ok := req.BasicAuth()
	if !ok {
		level.Warn(b.log).Log("msg", "invalid basic authentication information provided")
		return nil, errors.New("invalid basic authentication information")
	}
	return token.NewPrincipal(password, auth), nil
}

func (b *basicAuthPrincipalReader) CanProcess(req *http.Request) bool {
	auth := req.Header.Get("Authorization")
	return strings.HasPrefix(auth, "Basic ")
}

func BasicAuthPrincipalReader(logger log.Logger, verifier token.TokenVerifier) PrincipalReader {
	return &basicAuthPrincipalReader{
		log:      logger,
		verifier: verifier,
	}
}

// Bearer token Principal Reader
type bearTokenPrincipalReader struct {
	log      log.Logger
	verifier token.TokenVerifier
}

func (b *bearTokenPrincipalReader) Principal(req *http.Request) (token.IPrincipal, error) {
	auth := req.Header.Get("Authorization")
	if auth == "" {
		level.Info(b.log).Log("msg", "no token provided")
		return nil, errors.New("no Authentication Info")
	}

	tenant, password, ok := parseBearerToken(auth)
	if !ok {
		level.Warn(b.log).Log("msg", "invalid bearer token provided")
		return nil, errors.New("invalid bearer token information")
	}

	return token.NewPrincipal(password, tenant), nil
}

func (b *bearTokenPrincipalReader) CanProcess(req *http.Request) bool {
	auth := req.Header.Get("Authorization")
	return strings.HasPrefix(auth, "Bearer ")
}

func BearerTokenPrincipalReader(logger log.Logger, verifier token.TokenVerifier) PrincipalReader {
	return &bearTokenPrincipalReader{
		log:      logger,
		verifier: verifier,
	}
}

func parseBearerToken(auth string) (username, password string, ok bool) {
	const prefix = "Bearer "

	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

// Default Principal Reader
type defaultPrincipalReader struct {
	Header  string
	Tenants []string
}

func (d *defaultPrincipalReader) Principal(req *http.Request) (token.IPrincipal, error) {
	h := req.Header.Values(d.Header)
	if len(h) != 0 {
		return token.NewAuthorizedPrincipal(h...), nil
	}
	return token.NewAuthorizedPrincipal(d.Tenants...), nil
}

func (d *defaultPrincipalReader) CanProcess(req *http.Request) bool {
	auth := req.Header.Get("Authorization")
	return !strings.HasPrefix(auth, "Bearer ")
}

func HeaderPrincipalReader(header string, tenants ...string) PrincipalReader {
	return &defaultPrincipalReader{
		Header:  header,
		Tenants: tenants,
	}
}

// PrincipalChain looks for a principal in an array of principal getters and
// if it finds an error or a principal it returns, otherwise it returns (nil,nil).
type PrincipalChain []PrincipalReader

func (m PrincipalChain) Principal(r *http.Request) (token.IPrincipal, error) {
	for _, v := range m {
		if !v.CanProcess(r) {
			continue
		}

		p, err := v.Principal(r)
		if err != nil {
			return nil, err
		}

		if p != nil {
			return p, nil
		}
	}

	return nil, nil
}
