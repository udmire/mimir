package proxy

import (
	"github.com/gorilla/mux"
)

type ReverseProxyFactory interface {
	Register(router *mux.Router)
}

// type httpReverseProxyFactory struct {
// 	proxies []Proxy
// 	logger  log.Logger
// }

func HasComponent(cc ComponentProxyConfig) bool {
	return cc.Url == ""
}

// func NewReverseProxyFactory(cfg Config, registry routes.Registry, logger log.Logger) (ReverseProxyFactory, error) {
// 	componentProxyConfigs := []*ComponentProxyConfig{
// 		cfg.AdminApi.WithName(AdminApi),
// 		cfg.Distributor.WithName(Distributor),
// 		cfg.QueryFrontend.WithName(QueryFrontend),
// 		cfg.StoreGateway.WithName(StoreGateway),
// 		cfg.Ingester.WithName(Ingester),
// 	}
// 	if HasComponent(cfg.Ruler) {
// 		componentProxyConfigs = append(componentProxyConfigs, cfg.Ruler.WithName(Ruler))
// 	}
// 	if HasComponent(cfg.Alertmanager) {
// 		componentProxyConfigs = append(componentProxyConfigs, cfg.Alertmanager.WithName(AlertManager))
// 	}
// 	if HasComponent(cfg.Compactor) {
// 		componentProxyConfigs = append(componentProxyConfigs, cfg.Compactor.WithName(Compactor))
// 	}
// 	componentProxyConfigs = append(componentProxyConfigs, cfg.Default.WithName(Default))
//
// 	var proxies []Proxy
// 	for _, config := range componentProxyConfigs {
// 		proxy, err := NewComponentProxy(config, logger, registry.GetGroupRoutes(config.Name))
// 		if err != nil {
// 			level.Error(logger).Log("msg", "failed to build compoent reverse proxy", "name", config.Name, "err", err)
// 			return nil, err
// 		}
// 		proxies = append(proxies, proxy)
// 	}
//
// 	if cfg.InstanceConfig.Enabled {
// 		proxy, err := NewInstanceProxy(&cfg.InstanceConfig, logger)
// 		if err != nil {
// 			level.Error(logger).Log("msg", "failed to build instance reverse proxy", "err", err)
// 			return nil, err
// 		}
// 		proxies = append(proxies, proxy)
// 	}
//
// 	return &httpReverseProxyFactory{
// 		proxies: proxies,
// 		logger:  logger,
// 	}, nil
// }

// func (f *httpReverseProxyFactory) Register(router *mux.Router) {
// 	for _, proxy := range f.proxies {
// 		proxy.RegisterRoute(router)
// 	}
// }
