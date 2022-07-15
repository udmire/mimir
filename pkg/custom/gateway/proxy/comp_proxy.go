package proxy

import (
	"net/http"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/mimir/pkg/custom/utils/routes"
)

type componentProxy struct {
	cfg *ComponentProxyConfig

	logger     log.Logger
	proxy      ReverseProxy
	compRoutes []routes.Route
}

func NewComponentProxy(cfg *ComponentProxyConfig, logger log.Logger, routes []routes.Route) (Proxy, error) {
	proxy, err := NewProxy(cfg)
	if err != nil {
		return nil, err
	}
	return &componentProxy{
		cfg:        cfg,
		logger:     log.With(logger, "comp", cfg.Name),
		proxy:      proxy,
		compRoutes: routes,
	}, nil
}

func (c *componentProxy) HandlerFunc() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		c.proxy.Proxy(c.logger, rw, req)
	}
}

func (c *componentProxy) RegisterRoute(router *mux.Router) {
	if len(c.compRoutes) == 0 {
		return
	}

	for _, route := range c.compRoutes {
		route.Register(router, c.HandlerFunc())
	}
}
