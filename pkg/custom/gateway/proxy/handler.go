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

type ReverseProxyWrapper interface {
	Get() *httputil.ReverseProxy
	Matches(path string) bool
}

type httpReverseProxyWrapper struct {
	proxy    *httputil.ReverseProxy
	matchers utils.Matcher
}

func (h *httpReverseProxyWrapper) Get() *httputil.ReverseProxy {
	return h.proxy
}

func (h *httpReverseProxyWrapper) Matches(path string) bool {
	return h.matchers.Matches(path)
}

type ReverseProxyFactory interface {
	GetReverseProxy(path string) (*httputil.ReverseProxy, error)
}

type httpReverseProxyFactory struct {
	WrapperChain []ReverseProxyWrapper
	logger       log.Logger
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

	if cfg.Ruler != nil {
		componentProxyConfigs = append(componentProxyConfigs, cfg.Ruler.WithName(routes.Ruler))
	}

	if cfg.Alertmanager != nil {
		componentProxyConfigs = append(componentProxyConfigs, cfg.Alertmanager.WithName(routes.AlertManager))
	}

	if cfg.Compactor != nil {
		componentProxyConfigs = append(componentProxyConfigs, cfg.Compactor.WithName(routes.Compactor))
	}

	// TODO process the instance proxy

	componentProxyConfigs = append(componentProxyConfigs, cfg.Default.WithName(routes.Default))

	for _, config := range componentProxyConfigs {
		wrapper, err := newReverseProxyWrapper(config, routes.GetComponentRoutes(config.Name))
		if err != nil {
			level.Error(logger).Log("msg", "failed to build reverse proxy", "err", err)
			return nil, err
		}
		chain = append(chain, wrapper)
	}

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

func (h *httpReverseProxyFactory) GetReverseProxy(path string) (*httputil.ReverseProxy, error) {
	for _, wrapper := range h.WrapperChain {
		if wrapper.Matches(path) {
			return wrapper.Get(), nil
		}
	}
	return nil, fmt.Errorf("invalid configuration for ReverseProxys")
}

func WithProxy(factory ReverseProxyFactory, logger log.Logger) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, request *http.Request) {
		uri := request.RequestURI
		proxy, err := factory.GetReverseProxy(uri)
		if err != nil {
			utils.JSONError(logger, rw, "Authentication required", http.StatusUnauthorized)
			return
		}
		principal := auth.GetPrincipal(request.Context())
		if principal != nil {
			principal.WrapRequest(request)
		}
		proxy.ServeHTTP(rw, request)
	})
}
