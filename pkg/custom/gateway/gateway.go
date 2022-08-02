package gateway

import (
	"flag"

	"github.com/go-kit/log"
	"github.com/grafana/mimir/pkg/api"
	"github.com/grafana/mimir/pkg/custom/admin"
	"github.com/grafana/mimir/pkg/custom/gateway/auth"
	"github.com/grafana/mimir/pkg/custom/gateway/auth/access"
	"github.com/grafana/mimir/pkg/custom/gateway/proxy"
	"github.com/grafana/mimir/pkg/custom/utils/routes"
	"github.com/prometheus/client_golang/prometheus"
)

// Config xxx
type Config struct {
	Proxy  proxy.Config       `yaml:"proxy" category:"advanced"`
	Tenant proxy.TenantConfig `yaml:"tenant" category:"advanced"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet, logger log.Logger) {
	c.Proxy.RegisterFlags(f, logger)
	c.Tenant.RegisterFlags(f)
}

type Gateway struct {
	cfg *Config

	authServer *auth.AuthServer
	proxies    []proxy.Proxy
	registry   routes.Registry
	reg        prometheus.Registerer
	logger     log.Logger
}

// NewGateway creates a new gateway server.
func NewGateway(cfg Config, authCfg auth.Config, client *admin.Client, reg prometheus.Registerer, logger log.Logger) (*Gateway, error) {
	registry := proxy.NewRegistry()
	proxy.Init(registry)

	var proxies []proxy.Proxy

	var eval access.Evaluator
	if cfg.Proxy.InstanceConfig.Enabled {
		eval = access.NewPrefixedPermissionEvaluator(proxy.DynamicInstancePrefix, logger, registry)

		instanceProxy, err := proxy.NewInstanceProxy(&cfg.Proxy.InstanceConfig, logger)
		if err != nil {
			return nil, err
		}
		proxies = append(proxies, instanceProxy)
	} else {
		eval = access.NewPermissionEvaluator(logger, registry)
	}

	if cfg.Tenant.Enabled {
		tenantsHandler := proxy.NewTenantsPushProxy(&cfg.Proxy.Distributor, &cfg.Tenant, logger)
		proxies = append(proxies, tenantsHandler)
	}

	componentsProxy, err := proxy.NewComponentsProxy(cfg.Proxy, registry, logger)
	if err != nil {
		return nil, err
	}
	proxies = append(proxies, componentsProxy)

	authServer, err := auth.NewAuthServer(authCfg, eval, client, logger)
	if err != nil {
		return nil, err
	}

	return &Gateway{
		cfg:        &cfg,
		authServer: authServer,
		proxies:    proxies,
		registry:   registry,
		reg:        reg,
		logger:     logger,
	}, nil
}

func (g *Gateway) RegisterAPI(a *api.API) {
	a.AuthMiddleware = g.authServer

	g.registry.RegisterGroupLinks(a.RegisterComponent)

	for _, proxy := range g.proxies {
		method, additional := proxy.Methods()
		a.RegisterRoute(proxy.Path(), proxy.Handler(), true, false, method, additional...)
	}
}
