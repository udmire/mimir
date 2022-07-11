package proxy

import (
	"net/http/httputil"
	"net/url"
)

func NewProxy(cfg *ComponentProxyConfig) (*httputil.ReverseProxy, error) {
	remote, err := url.Parse(cfg.Url)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	return proxy, nil
}
