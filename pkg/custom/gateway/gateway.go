package gateway

import (
	"flag"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/mimir/pkg/custom/admin"
	"github.com/grafana/mimir/pkg/custom/gateway/auth"
	"github.com/grafana/mimir/pkg/custom/gateway/auth/access"
	"github.com/grafana/mimir/pkg/custom/gateway/proxy"
	"github.com/prometheus/client_golang/prometheus"
)

// Config xxx
type Config struct {
	Proxy proxy.Config `yaml:"proxy" category:"advanced"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet, logger log.Logger) {
	c.Proxy.RegisterFlags(f, logger)
}

type Gateway struct {
	cfg *Config

	authServer *auth.AuthServer
	proxies    []proxy.Proxy

	registry prometheus.Registerer
	logger   log.Logger
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
		registry:   reg,
		logger:     logger,
	}, nil
}

func (g *Gateway) Register(router *mux.Router) {
	// router.Use(func(handler http.Handler) http.Handler {
	// 	return auth.WithAuthentication(handler, g.authServer)
	// })

	for _, proxy := range g.proxies {
		proxy.RegisterRoute(router)
	}
}
