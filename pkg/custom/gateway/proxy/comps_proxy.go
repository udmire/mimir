package proxy

import (
	"net/http"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/mimir/pkg/custom/utils/routes"
	util_log "github.com/grafana/mimir/pkg/util/log"
)

const (
	ComponentRoute = "/{path:.+}"
)

type ComponentConfigFactory interface {
	GetComponentConfig(req *http.Request) *ComponentProxyConfig
}

type configFactory struct {
	Default *ComponentProxyConfig

	Components []*ComponentProxyConfig
	Registry   routes.Registry
}

func (c *configFactory) GetComponentConfig(req *http.Request) *ComponentProxyConfig {
	for _, component := range c.Components {
		permissions := c.Registry.GetGroupRoutesPermissions(component.Name)
		for _, permission := range permissions {
			if permission.Matches(req) {
				return component
			}
		}
	}
	return c.Default
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

func (c *compsProxy) RegisterRoute(router *mux.Router) {
	router.Handle(c.Path(), c.Handler())
}

func (c *compsProxy) Handler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		logger := util_log.WithContext(req.Context(), c.logger)
		config := c.GetComponentConfig(req)
		proxy, err := NewProxy(logger, config)
		if err != nil {
			return
		}
		proxy.Proxy(logger, rw, req)
	})
}

func (c *compsProxy) Path() string {
	return ComponentRoute
}
