package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Route interface {
	Methods() []string
	Pattern() string
	Register(router *mux.Router, hf http.HandlerFunc)
}

func WithPattern(pattern string) Route {
	return &internalRoute{
		methods: []string{"*"},
		pattern: pattern,
	}
}

func WithPatternMethods(pattern string, methods ...string) Route {
	return &internalRoute{
		methods: methods,
		pattern: pattern,
	}
}

type internalRoute struct {
	methods []string
	pattern string
}

func (i *internalRoute) Methods() []string {
	return i.methods
}

func (i *internalRoute) Pattern() string {
	return i.pattern
}

func (i *internalRoute) Register(router *mux.Router, hf http.HandlerFunc) {
	route := router.NewRoute()
	if len(i.methods) == 1 && i.methods[0] != "*" {
		route.Methods(i.methods...)
	}

	route.Path(precessPattern(i.pattern)).HandlerFunc(hf)
}

func precessPattern(pattern string) string {
	counter := 0
	for {
		if !strings.Contains(pattern, "/*") {
			break
		}
		pattern = strings.Replace(pattern, "/*", fmt.Sprintf("/{param%d}", counter), 1)
		counter++
	}
	if strings.HasSuffix(pattern, "**") {
		pattern = strings.Replace(pattern, "**", fmt.Sprintf("{param%d:.+}", counter), 1)
	}
	return pattern
}
