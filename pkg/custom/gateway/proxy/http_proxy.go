package proxy

import (
	"net/http"
	"net/url"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
)

type Proxy interface {
	HandlerFunc() http.HandlerFunc
	RegisterRoute(*mux.Router)
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

func NewDynamicProxy(logger log.Logger, target TargetFunc, path PathFunc) (ReverseProxy, error) {
	return NewHttpReverseProxy(logger, target, path), nil
}
