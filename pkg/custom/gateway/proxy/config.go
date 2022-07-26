package proxy

import (
	"flag"

	"github.com/go-kit/log"
	"github.com/grafana/dskit/crypto/tls"
	"github.com/grafana/dskit/flagext"
	"github.com/prometheus/common/model"
)

// Config for gateway proxy purpose
type Config struct {
	InstanceConfig InstanceProxyConfig  `yaml:"instance"`
	Default        ComponentProxyConfig `yaml:"default"`
	AdminApi       ComponentProxyConfig `yaml:"admin_api"`
	Alertmanager   ComponentProxyConfig `yaml:"alertmanager"`
	Compactor      ComponentProxyConfig `yaml:"compactor"`
	Distributor    ComponentProxyConfig `yaml:"distributor"`
	Ingester       ComponentProxyConfig `yaml:"ingester"`
	QueryFrontend  ComponentProxyConfig `yaml:"query_frontend"`
	Ruler          ComponentProxyConfig `yaml:"ruler"`
	StoreGateway   ComponentProxyConfig `yaml:"store_gateway"`
	Scraper        ComponentProxyConfig `yaml:"scraper"`
}

type ComponentProxyConfig struct {
	Url          string         `yaml:"url" category:"advanced"`
	Keepalive    bool           `yaml:"keepalive"`
	ReadTimeout  model.Duration `yaml:"read_timeout"`
	WriteTimeout model.Duration `yaml:"write_timeout"`

	TLSEnabled bool             `yaml:"tls_enabled" category:"advanced"`
	TLS        tls.ClientConfig `yaml:",inline"`

	Routes []RouteConfig `yaml:"routes,omitempty"`

	Name string `json:"-"`
}

type RouteConfig struct {
	Pattern     string                 `yaml:"pattern"`
	Methods     flagext.StringSliceCSV `yaml:"methods"`
	Permissions flagext.StringSliceCSV `yaml:"permissions"`
	Restrict    bool                   `yaml:"restrict"`
}

func (r *RouteConfig) RegisterFlags(prefix string, f *flag.FlagSet) {
	f.StringVar(&r.Pattern, prefix+".pattern", "", "Route Pattern")
	f.Var(&r.Methods, prefix+".methods", "List of pattern supported Method")
	f.Var(&r.Permissions, prefix+".permissions", "List of required permissions")
	f.BoolVar(&r.Restrict, prefix+".restrict", false, "Restrict to meet all the permissions")
}

func (c *ComponentProxyConfig) WithName(name string) *ComponentProxyConfig {
	c.Name = name
	return c
}

// RegisterFlagsWithPrefix registers flags with prefix.
func (c *ComponentProxyConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.StringVar(&c.Url, prefix+".url", "", "URL for the backend. Use the scheme dns:// for HTTP over GPRC and the scheme h2c:// for HTTP2 proxying.")
	f.Var(&c.ReadTimeout, prefix+".read-timeout", "Timeout for read requests the backend, set to <=0 to disable. (default 2m0s)")
	_ = c.ReadTimeout.Set("2m")
	f.Var(&c.WriteTimeout, prefix+".write-timeout", "Timeout for write requests to the backend, set to <=0 to disable. (default 30s)")
	_ = c.WriteTimeout.Set("30s")
	f.BoolVar(&c.Keepalive, prefix+".enable-keepalive", true, "Enable keep alive for the backend. (default true)")
	f.BoolVar(&c.TLSEnabled, prefix+".tls-enabled", c.TLSEnabled, "Enable TLS in the GRPC client. This flag needs to be enabled when any other TLS flag is set. If set to false, insecure connection to gRPC server will be used.")

	c.TLS.RegisterFlagsWithPrefix(prefix, f)
}

type InstanceProxyConfig struct {
	Enabled bool   `yaml:"enabled"`
	Pattern string `yaml:"pattern"`
}

func (c *InstanceProxyConfig) RegisterFlags(f *flag.FlagSet, logger log.Logger) {
	f.BoolVar(&c.Enabled, "gateway.proxy.instance.enabled", true, "Whether proxy for the whole app instances by name.")
	f.StringVar(&c.Pattern, "gateway.proxy.instance.pattern", "", "build target backend with the pattern.")
}

func (c *Config) RegisterFlags(f *flag.FlagSet, logger log.Logger) {
	c.InstanceConfig.RegisterFlags(f, logger)
	prefix := "gateway.proxy"
	c.Default.RegisterFlagsWithPrefix(prefix+".default", f)
	c.AdminApi.RegisterFlagsWithPrefix(prefix+".admin-api", f)
	c.Alertmanager.RegisterFlagsWithPrefix(prefix+".alertmanager", f)
	c.Compactor.RegisterFlagsWithPrefix(prefix+".compactor", f)
	c.Distributor.RegisterFlagsWithPrefix(prefix+".distributor", f)
	c.Ingester.RegisterFlagsWithPrefix(prefix+".ingester", f)
	c.QueryFrontend.RegisterFlagsWithPrefix(prefix+".query-frontend", f)
	c.Ruler.RegisterFlagsWithPrefix(prefix+".ruler", f)
	c.StoreGateway.RegisterFlagsWithPrefix(prefix+".store-gateway", f)
	c.Scraper.RegisterFlagsWithPrefix(prefix+".scraper", f)
}

type TenantConfig struct {
	Enabled        bool   `yaml:"enabled"`
	MatchType      string `yaml:"match_type"`
	MaxRecvMsgSize int    `yaml:"max_recv_msg_size"`
	TenantLabel    string `yaml:"tenant_label"`
	AcceptAll      bool   `yaml:"accept_all"`
}

func (t *TenantConfig) RegisterFlags(f *flag.FlagSet) {
	f.BoolVar(&t.Enabled, "gateway.tenant.enabled", false, "Enable mark label as tenant")
	f.StringVar(&t.MatchType, "gateway.tenant.match-type", "header", "Only support request with given header name")
	f.IntVar(&t.MaxRecvMsgSize, "gateway.tenant.max-recv-msg-size", 100<<20, "Max message size.")
	f.StringVar(&t.TenantLabel, "gateway.tenant.tenant-label", "__tenant__", "Tenant label name.")
	f.BoolVar(&t.AcceptAll, "gateway.tenant.accept-all", false, "Accept all even some failed.")
}
