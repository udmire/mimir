package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	util_log "github.com/grafana/mimir/pkg/util/log"
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

func (c *instanceProxy) Methods() (string, []string) {
	return http.MethodGet, []string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace,
	}
}

func NewInstanceProxy(cfg *InstanceProxyConfig, logger log.Logger) (Proxy, error) {
	path := func(req *http.Request) string {
		vars := mux.Vars(req)

		path, exists := vars[DynamicPath]
		if !exists {
			path = "/"
		}

		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		return path
	}

	target := func(req *http.Request) *url.URL {
		vars := mux.Vars(req)
		instance, exists := vars[Instance]
		if !exists {
			return nil
		}
		target, err := url.Parse(fmt.Sprintf(cfg.Pattern, instance))
		if err != nil {
			return nil
		}
		return target
	}

	proxy, err := NewDynamicProxy(logger, target, path)
	if err != nil {
		return nil, err
	}

	return &instanceProxy{
		cfg:    cfg,
		logger: logger,
		proxy:  proxy,
	}, nil
}

func (c *instanceProxy) Handler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		logger := util_log.WithContext(req.Context(), c.logger)
		logger = log.With(logger, "instance", vars[Instance])
		c.proxy.Proxy(logger, rw, req)
	})
}

func (c *instanceProxy) RegisterRoute(router *mux.Router) {
	router.Handle(DynamicInstanceRoute, c.Handler())
}

func (c *instanceProxy) Path() string {
	return DynamicInstanceRoute
}
