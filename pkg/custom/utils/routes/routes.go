package routes

import (
	"net/http"

	"github.com/grafana/mimir/pkg/custom/utils"
	"k8s.io/utils/strings/slices"
)

type Route interface {
	Methods() []string
	Pattern() string
	Matches(req *http.Request) bool
}

func WithPattern(pattern string) Route {
	return &internalRoute{
		methods: []string{},
		pattern: utils.MustCompile(pattern),
	}
}

func WithPatternMethods(pattern string, methods ...string) Route {
	return &internalRoute{
		methods: methods,
		pattern: utils.MustCompile(pattern),
	}
}

type internalRoute struct {
	methods []string
	pattern *utils.AntPattern
}

func (i *internalRoute) Methods() []string {
	return i.methods
}

func (i *internalRoute) Pattern() string {
	return i.pattern.String()
}

func (i *internalRoute) Matches(req *http.Request) bool {
	if len(i.methods) > 0 {
		if !slices.Contains(i.methods, req.Method) {
			return false
		}
	}
	return i.pattern.Matches(req.URL.Path)
}
