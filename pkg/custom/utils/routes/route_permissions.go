package routes

import (
	"fmt"
	"net/http"

	"github.com/grafana/mimir/pkg/custom/utils"
	"github.com/grafana/mimir/pkg/custom/utils/token"
)

type RoutePermissions interface {
	GetRoute() Route
	Matches(req *http.Request) bool
	HasPermission(principal token.IPrincipal) bool

	CopyWithPrefix(prefix string) RoutePermissions
}

func With(route Route, permissions ...string) RoutePermissions {
	return &internalRoutePermissions{
		route:       route,
		pattern:     utils.MustCompile(route.Path()),
		strict:      false,
		permissions: permissions,
	}
}

func StrictWith(route Route, permissions ...string) RoutePermissions {
	return &internalRoutePermissions{
		route:       route,
		pattern:     utils.MustCompile(route.Path()),
		strict:      false,
		permissions: permissions,
	}
}

type internalRoutePermissions struct {
	route       Route
	pattern     *utils.AntPattern
	permissions []string
	strict      bool
}

func (i *internalRoutePermissions) Matches(req *http.Request) bool {
	return i.route.Matches(req)
}

func (i *internalRoutePermissions) HasPermission(principal token.IPrincipal) bool {
	if i.strict {
		return principal.HasScopes(i.permissions...)
	} else {
		return principal.HasAnyScope(i.permissions...)
	}
}

func (i *internalRoutePermissions) GetRoute() Route {
	return i.route
}

func (i *internalRoutePermissions) CopyWithPrefix(prefix string) RoutePermissions {
	prefixPattern := fmt.Sprintf("%s%s", prefix, i.pattern)
	var methods []string
	method, additional := i.route.Methods()
	if method != "" {
		methods = append(methods, method)
	}
	if len(additional) > 0 {
		methods = append(methods, additional...)
	}
	route := WithAuthGzipPatternMethods(prefixPattern, i.route.Auth(), i.route.Gzip(), methods...)
	if i.strict {
		return StrictWith(route, i.permissions...)
	}

	return With(route, i.permissions...)
}
