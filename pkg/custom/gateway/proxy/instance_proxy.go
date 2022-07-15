package proxy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
)

const (
	DynamicInstancePrefix = "/dynamic/*"
	DynamicInstanceRoute  = "/dynamic/{instance}/{path:.*}"
	DynamicPath           = "path"
	Instance              = "instance"
)

type instanceProxy struct {
	cfg *InstanceProxyConfig

	logger log.Logger
	proxy  ReverseProxy
}

func NewInstanceProxy(cfg *InstanceProxyConfig, logger log.Logger) (Proxy, error) {
	modifier := func(req *http.Request) {
		vars := mux.Vars(req)

		path, exists := vars[DynamicPath]
		if !exists {
			path = "/"
		}

		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		req.URL.Path = path
	}

	target := func(instance string) string {
		return fmt.Sprintf(cfg.Pattern, instance)
	}

	proxy, err := NewDynamicProxy(target, modifier)
	if err != nil {
		return nil, err
	}

	return &instanceProxy{
		cfg:    cfg,
		logger: logger,
		proxy:  proxy,
	}, nil
}

func (c *instanceProxy) HandlerFunc() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		logger := log.With(c.logger, "instance", vars[Instance])
		c.proxy.Proxy(logger, rw, req)
	}
}

func (c *instanceProxy) RegisterRoute(router *mux.Router) {
	router.HandleFunc(DynamicInstanceRoute, c.HandlerFunc())
}
