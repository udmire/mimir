package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewProxy(cfg *ComponentProxyConfig) (ReverseProxy, error) {
	return NewHttpReverseProxy(func(req *http.Request) {}, func(req *http.Request) (*httputil.ReverseProxy, error) {
		remote, err := url.Parse(cfg.Url)
		if err != nil {
			return nil, err
		}
		return httputil.NewSingleHostReverseProxy(remote), nil
	}), nil
}

func NewDynamicProxy(target dynamicTarget, modifier requestModifier) (ReverseProxy, error) {
	return NewHttpReverseProxy(modifier, func(req *http.Request) (*httputil.ReverseProxy, error) {
		t := target(req.RequestURI)
		remote, err := url.Parse(t)
		if err != nil {
			return nil, err
		}
		return httputil.NewSingleHostReverseProxy(remote), nil
	}), nil
}
