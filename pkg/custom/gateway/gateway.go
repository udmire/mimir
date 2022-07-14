package gateway

import (
	"flag"

	"github.com/go-kit/log"
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
	cfg Config

	registry prometheus.Registerer
	logger   log.Logger
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

	return g, nil
}
