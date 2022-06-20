package gateway

import (
	"context"
	"flag"

	"github.com/go-kit/log"
	"github.com/grafana/dskit/services"
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
	services.Service

	cfg Config

	subservices        *services.Manager
	subservicesWatcher *services.FailureWatcher

	registry prometheus.Registerer
	logger   log.Logger
}

func (g *Gateway) starting(ctx context.Context) error {
	return services.StartManagerAndAwaitHealthy(ctx, g.subservices)
}

func (g *Gateway) run(ctx context.Context) error {
	return nil
}

func (g *Gateway) stopping(err error) error {
	return services.StopManagerAndAwaitStopped(context.Background(), g.subservices)
}

// NewGateway creates a new gateway server.
func NewGateway(cfg Config, reg prometheus.Registerer, logger log.Logger) (*Gateway, error) {
	return newGateway(cfg, reg, logger)
}

func newGateway(cfg Config, reg prometheus.Registerer, logger log.Logger) (g *Gateway, err error) {
	g = &Gateway{
		cfg:      cfg,
		registry: reg,
		logger:   logger,
	}
	// workers := &ProxyWorkers{
	// 	target:  "",
	// 	workers: map[string]*proxyWorker{},
	// }
	subservices := []services.Service(nil)

	g.subservices, err = services.NewManager(subservices...)
	if err != nil {
		return nil, err
	}
	g.subservicesWatcher = services.NewFailureWatcher()
	g.subservicesWatcher.WatchManager(g.subservices)

	g.Service = services.NewBasicService(g.starting, g.run, g.stopping)
	return g, nil
}

type ProxyWorkers struct {
	target  string
	workers map[string]*proxyWorker
}

func (p *ProxyWorkers) AddressAdded(address string) {
	p.workers[address] = &proxyWorker{}
}

func (p *ProxyWorkers) AddressRemoved(address string) {
	p.workers[address] = nil
}

type proxyWorker struct {
}
