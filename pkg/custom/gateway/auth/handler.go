package auth

import (
	"net/http"

	"github.com/go-kit/log/level"
	"github.com/grafana/mimir/pkg/custom/utils"
)

func (s *AuthServer) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		principal, err := s.authChain.Principal(r)
		if err != nil {
			level.Error(s.logger).Log("msg", "failed to get principal")
		}

		if principal == nil {
			utils.JSONError(s.logger, rw, "Authentication required", http.StatusUnauthorized)
			return
		}

		if !principal.IsAuthenticated() {
			err = principal.Verify(s.verifier)
			if err != nil {
				utils.JSONError(s.logger, rw, "Invalid Authentication", http.StatusUnauthorized)
				return
			}
			err = principal.LoadContext(s.loader)
			if err != nil {
				utils.JSONError(s.logger, rw, "Invalid Token Content", http.StatusUnauthorized)
				return
			}
			principal.Authenticate()
		}

		if !principal.IsAuthenticated() {
			utils.JSONError(s.logger, rw, "Authentication failed", http.StatusUnauthorized)
			return
		}

		canAccess := s.evaluator.Evaluate(r, principal)
		if !canAccess {
			utils.JSONError(s.logger, rw, "Access Denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(rw, r.Clone(SetPrincipal(r.Context(), principal)))
	})
}
