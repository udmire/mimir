package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/grafana/mimir/pkg/custom/utils"
	"k8s.io/utils/strings/slices"
)

type RouteMatcher interface {
	Matches(req *http.Request) bool
}

type Route interface {
	RouteMatcher
	Auth() bool
	Gzip() bool
	Methods() (method string, additional []string)
	Path() string
	Pattern() string
}

type ComponentRoute interface {
	Routes() []Route
}

func WithPattern(pattern string) Route {
	return WithPatternMethods(pattern)
}
func WithPatternMethods(pattern string, methods ...string) Route {
	return WithAuthPatternMethods(pattern, false, methods...)
}
func WithAuthPatternMethods(pattern string, auth bool, methods ...string) Route {
	return WithAuthGzipPatternMethods(pattern, auth, false, methods...)
}
func WithGzipPatternMethods(pattern string, gzip bool, methods ...string) Route {
	return WithAuthGzipPatternMethods(pattern, false, gzip, methods...)
}
func WithAuthGzipPatternMethods(pattern string, auth, gzip bool, methods ...string) Route {
	var method string
	var additional []string
	if len(methods) > 0 {
		method = methods[0]
		additional = methods[0:]
	}
	return &internalRoute{
		method:            method,
		additionalMethods: additional,
		auth:              auth,
		gzip:              gzip,
		pattern:           utils.MustCompile(pattern),
	}
}

type internalRoute struct {
	method            string
	additionalMethods []string
	auth              bool
	gzip              bool
	pattern           *utils.AntPattern
}

func (i *internalRoute) Matches(req *http.Request) bool {
	if len(i.method) > 0 {
		if i.method != req.Method && !slices.Contains(i.additionalMethods, req.Method) {
			return false
		}
	}
	return i.pattern.Matches(req.URL.Path)
}

func (i *internalRoute) Auth() bool {
	return i.auth
}

func (i *internalRoute) Gzip() bool {
	return i.gzip
}

func (i *internalRoute) Methods() (method string, additional []string) {
	return i.method, i.additionalMethods
}

func (i *internalRoute) Path() string {
	return i.pattern.String()
}

func (i *internalRoute) Pattern() string {
	return precessPattern(i.pattern.String())
}

func precessPattern(pattern string) string {
	counter := 0
	for {
		if !(strings.Contains(pattern, "/*") && strings.Contains(pattern, "*/")) {
			break
		}

		if strings.Contains(pattern, "/**/") {
			pattern = strings.Replace(pattern, "/**/", fmt.Sprintf("/{param%d:.+}/", counter), 1)
		} else if strings.Contains(pattern, "/*/") {
			pattern = strings.Replace(pattern, "/*/", fmt.Sprintf("/{param%d}/", counter), 1)
		}
		counter++
	}
	if strings.HasSuffix(pattern, "**") {
		pattern = strings.Replace(pattern, "**", fmt.Sprintf("{param%d:.+}", counter), 1)
	}
	return pattern
}
