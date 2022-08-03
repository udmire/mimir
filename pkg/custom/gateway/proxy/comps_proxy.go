package proxy

import (
	"net/http"

	"github.com/go-kit/log"
	"github.com/grafana/mimir/pkg/custom/utils/routes"
	util_log "github.com/grafana/mimir/pkg/util/log"
)

const (
	DefaultComponentRoute = "/{all:.+}"
)

type configFactory struct {
	Default *ComponentProxyConfig

	Components []*ComponentProxyConfig
	Registry   routes.Registry
}

type compsProxy struct {
	configFactory

	logger log.Logger
}

func NewComponentsProxy(cfg Config, registry routes.Registry, logger log.Logger) (*compsProxy, error) {

	componentProxyConfigs := []*ComponentProxyConfig{
		cfg.AdminApi.WithName(AdminApi),
		cfg.Distributor.WithName(Distributor),
		cfg.QueryFrontend.WithName(QueryFrontend),
		cfg.StoreGateway.WithName(StoreGateway),
		cfg.Ingester.WithName(Ingester),
	}
	if HasComponent(cfg.Ruler) {
		componentProxyConfigs = append(componentProxyConfigs, cfg.Ruler.WithName(Ruler))
	}
	if HasComponent(cfg.Alertmanager) {
		componentProxyConfigs = append(componentProxyConfigs, cfg.Alertmanager.WithName(AlertManager))
	}
	if HasComponent(cfg.Compactor) {
		componentProxyConfigs = append(componentProxyConfigs, cfg.Compactor.WithName(Compactor))
	}

	return &compsProxy{
		logger: logger,
		configFactory: configFactory{
			Default:    cfg.Default.WithName(Default),
			Components: componentProxyConfigs,
			Registry:   registry,
		},
	}, nil
}

func (c *compsProxy) Handler(config *ComponentProxyConfig) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		logger := util_log.WithContext(req.Context(), c.logger)
		rewrites := c.Registry.GetGroupRewrites(config.Name)
		var proxy ReverseProxy
		var err error
		if len(rewrites) > 0 {
			proxy, err = NewRewriteProxy(logger, config, rewrites)
		} else {
			proxy, err = NewProxy(logger, config)
		}
		if err != nil {
			return
		}
		proxy.Proxy(logger, rw, req)
	})
}

func (c *compsProxy) RegisterRoutes(f func(path string, handler http.Handler, auth bool, gzipEnabled bool, method string, methods ...string)) {
	for _, component := range c.Components {
		groupRoutes := c.Registry.GetGroupRoutes(component.Name)
		for _, route := range groupRoutes {
			method, additional := route.Methods()
			if method == "" {
				method = defaultMethod
				additional = defaultAdditionalMethods
			}
			f(route.Path(), c.Handler(component), route.Auth(), route.Gzip(), method, additional...)
		}
	}

	f(DefaultComponentRoute, c.Handler(c.Default), true, true, defaultMethod, defaultAdditionalMethods...)
}
