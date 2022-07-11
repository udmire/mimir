package auth

import (
	"net/http"

	"github.com/go-kit/log/level"
	"github.com/grafana/mimir/pkg/custom/utils"
)

// WithAuth middleware adds auth validation to API handlers.
//
// Unauthorized requests will be denied with a 401 status code.
func WithAuth(next http.Handler, srv *AuthServer) http.Handler {
	basicAuth := BasicAuthPrincipalReader(srv.logger, srv.verifier)
	chain := PrincipalChain{basicAuth}

	if srv.OidcEnabled() {
		bearerTokenAuth := BearerTokenPrincipalReader(srv.logger, srv.verifier)
		chain = append(chain, bearerTokenAuth)
	}

	routes := srv.GetPublicRoutes()
	matchers := NewPublicRouteMatchers(routes)

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if matchers.Match(r) {
			next.ServeHTTP(rw, r)
			return
		}

		principal, err := chain.Principal(r)
		if err != nil {
			level.Error(srv.logger).Log("msg", "failed to get principal")
		}

		if principal == nil || err != nil {
			utils.JSONError(srv.logger, rw, "Authentication required", http.StatusUnauthorized)
			return
		}

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

		canAccess := srv.evaluator.Evaluate(r, principal)
		if !canAccess {
			utils.JSONError(srv.logger, rw, "Access Denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(rw, r.Clone(SetPrincipal(r.Context(), principal)))
	})
}
