package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/mimir/pkg/custom/auth"
	"github.com/grafana/mimir/pkg/custom/utils"
	"github.com/grafana/mimir/pkg/custom/utils/routes"
	_ "github.com/grafana/mimir/pkg/custom/utils/routes"
)

type ReverseProxy interface {
	Proxy(logger log.Logger, rw http.ResponseWriter, request *http.Request)
}

type reverseProxy struct {
	modifier requestModifier
	proxy    func(req *http.Request) (*httputil.ReverseProxy, error)
}

func (r *reverseProxy) Proxy(logger log.Logger, rw http.ResponseWriter, request *http.Request) {
	r.modifier(request)
	proxy, err := r.proxy(request)
	if err != nil {
		utils.JSONError(logger, rw, "", http.StatusInternalServerError)
		return
	}
	proxy.ServeHTTP(rw, request)
}

func NewHttpReverseProxy(modifier requestModifier, proxy func(req *http.Request) (*httputil.ReverseProxy, error)) ReverseProxy {
	return &reverseProxy{
		proxy:    proxy,
		modifier: modifier,
	}
}

type ReverseProxyWrapper interface {
	Get() ReverseProxy
	Matches(path string) bool
}

type httpReverseProxyWrapper struct {
	proxy    ReverseProxy
	matchers utils.Matcher
}

func (h *httpReverseProxyWrapper) Get() ReverseProxy {
	return h.proxy
}

func (h *httpReverseProxyWrapper) Matches(path string) bool {
	return h.matchers.Matches(path)
}

type ReverseProxyFactory interface {
	GetReverseProxy(path string) (ReverseProxy, error)
}

type httpReverseProxyFactory struct {
	WrapperChain []ReverseProxyWrapper
	logger       log.Logger
}

func HasComponent(cc ComponentProxyConfig) bool {
	return cc == ComponentProxyConfig{} || cc.Url == ""
}

func NewReverseProxyFactory(cfg Config, logger log.Logger) (ReverseProxyFactory, error) {
	var chain []ReverseProxyWrapper

	componentProxyConfigs := []*ComponentProxyConfig{
		cfg.AdminApi.WithName(routes.AdminApi),
		cfg.Distributor.WithName(routes.Distributor),
		cfg.QueryFrontend.WithName(routes.QueryFrontend),
		cfg.StoreGateway.WithName(routes.StoreGateway),
		cfg.Ingester.WithName(routes.Ingester),
	}

	if HasComponent(cfg.Ruler) {
		componentProxyConfigs = append(componentProxyConfigs, cfg.Ruler.WithName(routes.Ruler))
	}

	if HasComponent(cfg.Alertmanager) {
		componentProxyConfigs = append(componentProxyConfigs, cfg.Alertmanager.WithName(routes.AlertManager))
	}

	if HasComponent(cfg.Compactor) {
		componentProxyConfigs = append(componentProxyConfigs, cfg.Compactor.WithName(routes.Compactor))
	}

	for _, config := range componentProxyConfigs {
		wrapper, err := newReverseProxyWrapper(config, routes.GetComponentRoutes(config.Name))
		if err != nil {
			level.Error(logger).Log("msg", "failed to build reverse proxy", "err", err)
			return nil, err
		}
		chain = append(chain, wrapper)
	}

	instanceProxy, err := newDynamicInstanceProxy(cfg.InstanceConfig, routes.GetComponentRoutes(routes.Instance))
	if err != nil {
		level.Error(logger).Log("msg", "failed to build instance reverse proxy", "err", err)
		return nil, err
	}
	if instanceProxy != nil {
		chain = append(chain, instanceProxy)
	}

	wrapper, err := newReverseProxyWrapper(cfg.Default.WithName(routes.Default), routes.GetComponentRoutes(routes.Default))
	if err != nil {
		level.Error(logger).Log("msg", "failed to build default reverse proxy", "err", err)
		return nil, err
	}
	chain = append(chain, wrapper)

	return &httpReverseProxyFactory{
		WrapperChain: chain,
		logger:       logger,
	}, nil
}

func newReverseProxyWrapper(config *ComponentProxyConfig, route *routes.ComponentRoutes) (ReverseProxyWrapper, error) {
	proxy, err := NewProxy(config)
	if err != nil {
		return nil, err
	}
	return &httpReverseProxyWrapper{
		proxy:    proxy,
		matchers: route.HttpRoutes,
	}, nil
}

func (h *httpReverseProxyFactory) GetReverseProxy(path string) (ReverseProxy, error) {
	if h.noProxyPath(path) {
		return nil, nil
	}

	for _, wrapper := range h.WrapperChain {
		if wrapper.Matches(path) {
			return wrapper.Get(), nil
		}
	}
	return nil, fmt.Errorf("invalid configuration for ReverseProxys")
}

func (h *httpReverseProxyFactory) noProxyPath(path string) bool {
	return "/metrics" == path
}

func WithProxy(next http.Handler, factory ReverseProxyFactory, logger log.Logger) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, request *http.Request) {
		// TODO
		requestLogger := log.With(logger, "request", request.Context().Value("trace"))
		uri := request.RequestURI
		proxy, err := factory.GetReverseProxy(uri)
		if err != nil {
			utils.JSONError(logger, rw, "Authentication required", http.StatusUnauthorized)
			return
		}

		if proxy == nil {
			next.ServeHTTP(rw, request)
			return
		}

		principal := auth.GetPrincipal(request.Context())
		if principal != nil {
			principal.WrapRequest(request)
		}
		proxy.Proxy(requestLogger, rw, request)
	})
}
