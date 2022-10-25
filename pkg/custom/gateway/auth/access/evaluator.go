package access

import (
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
	allRoutePermissions []routes.RoutePermissions
}

func NewPermissionEvaluator(logger log.Logger, registry routes.Registry) Evaluator {
	return &permissionEvaluator{
		logger:              logger,
		allRoutePermissions: registry.GetAllRoutesPermissions(),
	}
}

func NewPrefixedPermissionEvaluator(prefix string, logger log.Logger, registry routes.Registry) Evaluator {
	var all []routes.RoutePermissions
	permissions := registry.GetAllRoutesPermissions()
	for _, p := range permissions {
		all = append(all, p.CopyWithPrefix(prefix))
	}
	return &permissionEvaluator{
		logger:              logger,
		allRoutePermissions: append(all, permissions...),
	}
}

func (p *permissionEvaluator) Evaluate(req *http.Request, principal token.IPrincipal) bool {
	for _, permission := range p.allRoutePermissions {
		if permission.Matches(req) {
			return permission.HasPermission(principal)
		}
	}

	return false
}
