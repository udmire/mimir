package auth

import (
	"net/http"

	"github.com/grafana/mimir/pkg/custom/utils"
)

type RouteMatcher interface {
	Match(req *http.Request) bool
}

type publicRouteMatchers struct {
	matchers []*utils.AntPattern
}

func NewPublicRouteMatchers(publicRoutes []string) *publicRouteMatchers {
	var matchers []*utils.AntPattern
	if len(publicRoutes) > 0 {
		for _, route := range publicRoutes {
			compile := utils.MustCompile(route)
			matchers = append(matchers, compile)
		}
	}
	return &publicRouteMatchers{
		matchers: matchers,
	}
}

func (p *publicRouteMatchers) Match(req *http.Request) bool {
	if len(p.matchers) == 0 {
		return false
	}
	uri := req.RequestURI
	for _, matcher := range p.matchers {
		if matcher.Matches(uri) {
			return true
		}
	}
	return false
}
