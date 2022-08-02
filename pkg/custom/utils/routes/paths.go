package routes

import (
	"net/http"

	"github.com/prometheus/prometheus/model/relabel"
)

type Rewrite interface {
	Rewrite(req *http.Request)
}

type Rewriter interface {
	Rewrite
	RouteMatcher
}

type RewriteFactory interface {
	Rewrite

	WithRouteAndRules(route Route, regex, replacement string) error
}

func RewriteFactoryWith() (RewriteFactory, error) {
	return &internalRewriteFactory{}, nil
}

type internalRewriteFactory struct {
	rewrites []Rewriter
}

func (i *internalRewriteFactory) WithRouteAndRules(route Route, regex, replacement string) error {
	rewriter, err := NewRewriter(route, regex, regex)
	if err != nil {
		return err
	}
	i.rewrites = append(i.rewrites, rewriter)
	return nil
}

func (i *internalRewriteFactory) Rewrite(req *http.Request) {
	for _, rewrite := range i.rewrites {
		if rewrite.Matches(req) {
			rewrite.Rewrite(req)
		}
	}
}

type internalRouteReplacer struct {
	routeMatcher RouteMatcher
	regexp       *relabel.Regexp
	replacement  string
}

func (i *internalRouteReplacer) Matches(req *http.Request) bool {
	return i.routeMatcher.Matches(req)
}

func NewRewriter(routerMatcher RouteMatcher, matcher, replacement string) (Rewriter, error) {
	if len(matcher) == 0 || len(replacement) == 0 {
		return nil, nil
	}

	regex, err := relabel.NewRegexp(matcher)
	if err != nil {
		return nil, err
	}

	return &internalRouteReplacer{
		routeMatcher: routerMatcher,
		regexp:       &regex,
		replacement:  replacement,
	}, nil
}

func (i *internalRouteReplacer) Rewrite(req *http.Request) {
	if i.regexp == nil {
		return
	}

	indexes := i.regexp.FindStringSubmatchIndex(req.URL.Path)
	if indexes == nil {
		return
	}

	res := i.regexp.ExpandString([]byte{}, i.replacement, req.URL.Path, indexes)
	if len(res) == 0 {
		return
	}
	req.URL.Path = string(res)
}
