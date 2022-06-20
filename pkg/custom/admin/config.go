package admin

import (
	"flag"
	"time"

	"github.com/go-kit/log"
)

type ApiLimitConfig struct {
	Enabled       bool          `yaml:"enabled" category:"advanced"`
	RefreshPeriod time.Duration `yaml:"refresh_period" category:"advanced"`
}

func (c *ApiLimitConfig) RegisterFlags(f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, "admin-api.limits.enabled", true, "Enable API based limits per-tenant.")
	f.DurationVar(&c.RefreshPeriod, "admin-api.limits.refresh-period", time.Minute, "Period with which to refresh per-tenant limits.")
}

type ApiConfig struct {
	UserHeaderName string               `yaml:"user_header_name" category:"advanced"`
	LeaderElection LeaderElectionConfig `yaml:"leader_election" category:"advanced"`
	ClusterName    string               `yaml:"cluster_name" category:"advanced"`
	Limits         ApiLimitConfig       `yaml:"limits" category:"advanced"`
}

func (c *ApiConfig) RegisterFlags(f *flag.FlagSet, logger log.Logger) {
	f.StringVar(&c.UserHeaderName, "admin.api.user-header-name", "X-WEBAUTH-USER", "Designated header to parse when searching for the grafana user ID of the user accessing the API.")
	f.StringVar(&c.ClusterName, "admin.api.cluster-name", "cortex", "Cluster Name")
	c.LeaderElection.RegisterFlags(f, logger)
	c.Limits.RegisterFlags(f)
}
