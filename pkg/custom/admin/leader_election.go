package admin

import (
	"flag"

	"github.com/go-kit/log"
	"github.com/grafana/dskit/grpcclient"
)

type LeaderElectionConfig struct {
	Enabled      bool              `yaml:"enabled" category:"advanced"`
	Ring         RingConfig        `yaml:"ring" category:"advanced"`
	ClientConfig grpcclient.Config `yaml:"client_config" category:"advanced"`
}

func (c *LeaderElectionConfig) RegisterFlags(f *flag.FlagSet, logger log.Logger) {
	f.BoolVar(&c.Enabled, "admin-api.leader-election.enabled", true, "This flag enables leader election for the admin api")
	c.Ring.RegisterFlags(f, logger)
	c.ClientConfig.RegisterFlagsWithPrefix("admin-api.leader-election.client", f)
}
