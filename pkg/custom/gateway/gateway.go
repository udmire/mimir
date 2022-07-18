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
	factory    proxy.ReverseProxyFactory

	registry prometheus.Registerer
	logger   log.Logger
}

// NewGateway creates a new gateway server.
func NewGateway(cfg Config, authCfg auth.Config, client *admin.Client, reg prometheus.Registerer, logger log.Logger) (*Gateway, error) {
	registry := NewRegistry()
	Init(registry)

	factory, err := proxy.NewReverseProxyFactory(cfg.Proxy, registry, logger)
	if err != nil {
		return nil, err
	}

	var eval access.Evaluator
	if cfg.Proxy.InstanceConfig.Enabled {
		eval = access.NewPrefixedPermissionEvaluator(proxy.DynamicInstancePrefix, logger, registry)
	} else {
		eval = access.NewPermissionEvaluator(logger, registry)
	}

	authServer, err := auth.NewAuthServer(authCfg, eval, client, logger)
	if err != nil {
		return nil, err
	}

	return &Gateway{
		cfg:        &cfg,
		authServer: authServer,
		factory:    factory,
		registry:   reg,
		logger:     logger,
	}, nil
}

func (g *Gateway) Register(router *mux.Router) {
	// router.Use(func(handler http.Handler) http.Handler {
	// 	return auth.WithAuthentication(handler, g.authServer)
	// })

	g.factory.Register(router)
}
