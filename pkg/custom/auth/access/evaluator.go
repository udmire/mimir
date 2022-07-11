package access

import (
	"fmt"
	"net/http"

	"github.com/go-kit/log"
	"github.com/grafana/mimir/pkg/custom/utils/routes"
	"github.com/grafana/mimir/pkg/custom/utils/token"
)

type Evaluator interface {
	Evaluate(req *http.Request, principal token.IPrincipal) bool
}

type permissionEvaluator struct {
	logger              log.Logger
	allRoutePermissions routes.AccessPermissions
}

func NewPermissionEvaluator(logger log.Logger) Evaluator {
	return &permissionEvaluator{
		logger:              logger,
		allRoutePermissions: routes.GetAllRoutesPermissions(),
	}
}

func (p *permissionEvaluator) Evaluate(req *http.Request, principal token.IPrincipal) bool {
	action := fmt.Sprintf("%s%s", req.Method, req.RequestURI)
	for _, permission := range p.allRoutePermissions {
		if permission.Matches(action) {
			return permission.HasPermission(principal)
		}
	}

	return false
}
