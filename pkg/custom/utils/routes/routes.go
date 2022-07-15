package routes

import (
	"net/http"

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
	if len(i.methods) == 1 && i.methods[0] != "*" {
		router.Methods(i.methods...)
	}

	router.Path(i.pattern).HandlerFunc(hf)
}
