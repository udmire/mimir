package auth

import (
	"net/http"

	"github.com/go-kit/log/level"
	"github.com/grafana/mimir/pkg/custom/utils"
)

// WithAuthentication middleware adds auth validation to API handlers.
// Unauthorized requests will be denied with a 401 status code.
func WithAuthentication(next http.Handler, srv *AuthServer) http.Handler {
	basicAuth := BasicAuthPrincipalReader(srv.logger, srv.verifier)
	chain := PrincipalChain{basicAuth}

	if srv.OidcEnabled() {
		bearerTokenAuth := BearerTokenPrincipalReader(srv.logger, srv.verifier)
		chain = append(chain, bearerTokenAuth)
	}

	if srv.HeaderEnabled() {
		headerAuth := HeaderPrincipalReader(srv.cfg.Admin.Header.HeaderName, srv.cfg.Admin.Header.DefaultTenants...)
		chain = append(chain, headerAuth)
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if srv.matchers.Match(r) {
			next.ServeHTTP(rw, r)
			return
		}

		principal, err := chain.Principal(r)
		if err != nil {
			level.Error(srv.logger).Log("msg", "failed to get principal")
		}

		if principal == nil {
			utils.JSONError(srv.logger, rw, "Authentication required", http.StatusUnauthorized)
			return
		}

		if !principal.IsAuthenticated() {
			err = principal.Verify(srv.verifier)
			if err != nil {
				utils.JSONError(srv.logger, rw, "Invalid Authentication", http.StatusUnauthorized)
				return
			}
			err = principal.LoadContext(srv.loader)
			if err != nil {
				utils.JSONError(srv.logger, rw, "Invalid Token Content", http.StatusUnauthorized)
				return
			}
			principal.Authenticate()
		}

		if !principal.IsAuthenticated() {
			utils.JSONError(srv.logger, rw, "Authentication failed", http.StatusUnauthorized)
			return
		}

		canAccess := srv.evaluator.Evaluate(r, principal)
		if !canAccess {
			utils.JSONError(srv.logger, rw, "Access Denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(rw, r.Clone(SetPrincipal(r.Context(), principal)))
	})
}
