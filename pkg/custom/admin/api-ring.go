package admin

import (
	"flag"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/flagext"
	"github.com/grafana/dskit/kv"
	"github.com/grafana/dskit/netutil"
	"github.com/grafana/dskit/ring"
	util_log "github.com/grafana/mimir/pkg/util/log"
)

const (
	// If a admin api is unable to heartbeat the ring, its better to quickly remove it.
	ringAutoForgetUnhealthyPeriods = 2
)

// RingConfig masks the ring lifecycler config which contains
// many options not really required by the admin leader election ring. This config
// is used to strip down the config to the minimum, and avoid confusion
// to the user.
type RingConfig struct {
	KVStore             kv.Config     `yaml:"kvstore"`
	HeartbeatPeriod     time.Duration `yaml:"heartbeat_period" category:"advanced"`
	HeartbeatTimeout    time.Duration `yaml:"heartbeat_timeout" category:"advanced"`
	TokensObservePeriod time.Duration `yaml:"tokens_observe_period" category:"advanced"`

	// Instance details
	InstanceID             string   `yaml:"instance_id" doc:"default=<hostname>" category:"advanced"`
	InstanceInterfaceNames []string `yaml:"instance_interface_names" doc:"default=[<private network interfaces>]"`
	InstancePort           int      `yaml:"instance_port" category:"advanced"`
	InstanceAddr           string   `yaml:"instance_addr" category:"advanced"`

	// Injected internally
	ListenPort int `yaml:"-"`
}

// RegisterFlags adds the flags required to config this to the given FlagSet
func (cfg *RingConfig) RegisterFlags(f *flag.FlagSet, logger log.Logger) {
	hostname, err := os.Hostname()
	if err != nil {
		level.Error(util_log.Logger).Log("msg", "failed to get hostname", "err", err)
		os.Exit(1)
	}

	// Ring flags
	cfg.KVStore.Store = "memberlist"
	cfg.KVStore.RegisterFlagsWithPrefix("admin-api.leader-election.ring.", "leader-election/", f)
	f.DurationVar(&cfg.HeartbeatPeriod, "admin-api.leader-election.ring.heartbeat-period", 15*time.Second, "Period at which to heartbeat to the ring.")
	f.DurationVar(&cfg.HeartbeatTimeout, "admin-api.leader-election.ring.heartbeat-timeout", time.Minute, "The heartbeat timeout after which admin-api instances are considered unhealthy within the ring")
	f.DurationVar(&cfg.TokensObservePeriod, "admin-api.leader-election.ring.tokens-observe-period", time.Minute, "Period to wait after generating tokens to resolve collisions. Required when using a gossip ring KV store.")

	// Instance flags
	cfg.InstanceInterfaceNames = netutil.PrivateNetworkInterfacesWithFallback([]string{"eth0", "en0"}, logger)
	f.Var((*flagext.StringSlice)(&cfg.InstanceInterfaceNames), "admin-api.leader-election.ring.instance-interface-names", "List of network interface names to look up when finding the instance IP address.")
	f.StringVar(&cfg.InstanceAddr, "admin-api.leader-election.ring.instance-addr", "", "IP address to advertise in the ring. Default is auto-detected.")
	f.IntVar(&cfg.InstancePort, "admin-api.leader-election.ring.instance-port", 0, "Port to advertise in the ring (defaults to -server.grpc-listen-port).")
	f.StringVar(&cfg.InstanceID, "admin-api.leader-election.ring.instance-id", hostname, "Instance ID to register in the ring.")
}

// ToLifecyclerConfig returns a LifecyclerConfig based on the admin api leader election ring config.
func (cfg *RingConfig) ToLifecyclerConfig() ring.LifecyclerConfig {
	lc := ring.LifecyclerConfig{}
	rc := ring.Config{}

	flagext.DefaultValues(&lc)
	flagext.DefaultValues(&rc)

	// Configure ring
	rc.KVStore = cfg.KVStore
	rc.HeartbeatTimeout = cfg.HeartbeatTimeout
	rc.ReplicationFactor = 1

	// Configure lifecycler
	lc.RingConfig = rc
	lc.ListenPort = cfg.ListenPort
	lc.Addr = cfg.InstanceAddr
	lc.Port = cfg.InstancePort
	lc.ID = cfg.InstanceID
	lc.InfNames = cfg.InstanceInterfaceNames
	lc.UnregisterOnShutdown = true
	lc.HeartbeatPeriod = cfg.HeartbeatPeriod
	lc.HeartbeatTimeout = cfg.HeartbeatTimeout
	lc.ObservePeriod = cfg.TokensObservePeriod
	lc.NumTokens = 1
	lc.JoinAfter = 0
	lc.MinReadyDuration = 0
	lc.FinalSleep = 0

	return lc
}

func (cfg *RingConfig) ToRingConfig() ring.Config {
	rc := ring.Config{}
	flagext.DefaultValues(&rc)

	rc.KVStore = cfg.KVStore
	rc.HeartbeatTimeout = cfg.HeartbeatTimeout
	rc.ReplicationFactor = 1

	return rc
}
