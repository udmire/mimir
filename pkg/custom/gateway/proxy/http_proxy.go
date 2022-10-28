package proxy

import (
	"net/http"
	"net/url"

	"github.com/go-kit/log"
	"github.com/grafana/mimir/pkg/custom/utils/routes"
)

var (
	defaultMethod            = http.MethodGet
	defaultAdditionalMethods = []string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace,
	}
)

type Proxy interface {
	RegisterRoutes(func(path string, handler http.Handler, auth bool, gzipEnabled bool, method string, methods ...string))
}

func NewProxy(logger log.Logger, cfg *ComponentProxyConfig) (ReverseProxy, error) {
	return NewHttpReverseProxy(logger, func(req *http.Request) *url.URL {
		remote, err := url.Parse(cfg.Url)
		if err != nil {
			return nil
		}
		return remote
	}, func(req *http.Request) string {
		return req.URL.Path
	}), nil
}

func NewRewriteProxy(logger log.Logger, cfg *ComponentProxyConfig, rewrites []routes.Rewriter) (ReverseProxy, error) {
	return NewHttpReverseProxy(logger, func(req *http.Request) *url.URL {
		remote, err := url.Parse(cfg.Url)
		if err != nil {
			return nil
		}
		return remote
	}, func(req *http.Request) string {
		for _, rewrite := range rewrites {
			if rewrite.Matches(req) {
				rewrite.Rewrite(req)
				break
			}
		}
		return req.URL.Path
	}), nil
}

func NewDynamicProxy(logger log.Logger, target TargetFunc, path PathFunc) (ReverseProxy, error) {
	return NewHttpReverseProxy(logger, target, path), nil
}