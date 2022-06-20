package admin

import (
	"context"
	"flag"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/mimir/pkg/custom/admin/store"
	"github.com/grafana/mimir/pkg/storage/bucket"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	DefaultAdminPolicyName = "__admin__"
)

type StorageConfig struct {
	bucket.Config `yaml:",inline"`
	EnableCache   bool `yaml:"enable_cache" category:"advanced"`
}

// RegisterFlags registers the backend storage config.
func (c *StorageConfig) RegisterFlags(f *flag.FlagSet) {
	prefix := "admin.client."
	c.Filesystem.RegisterFlagsWithPrefixAndDefaultDirectory(prefix, "admin-api", f)
	f.BoolVar(&c.EnableCache, prefix+"cache.enabled", true, "Enable caching on the versioned client.")
}

type ClientConfig struct {
	DisableDefaultAdminPolicy bool          `yaml:"disable_default_admin_policy" category:"advanced"`
	Storage                   StorageConfig `yaml:"storage"`
}

func (c *ClientConfig) RegisterFlags(f *flag.FlagSet) {
	f.BoolVar(&c.DisableDefaultAdminPolicy, "admin.client.disable-default-admin-policy", false, "If set to true, the built-in __admin__ access policy will not be active.")
	c.Storage.RegisterFlags(f)
}

type Client struct {
	cfg    ClientConfig
	store  store.ApiStore
	logger log.Logger
}

func NewClient(ctx context.Context, cfg ClientConfig, logger log.Logger, reg prometheus.Registerer) (*Client, error) {
	storeBucket, err := NewApiStoreBucket(ctx, cfg.Storage, logger, reg)
	if err != nil {
		return nil, err
	}

	return &Client{
		cfg:    cfg,
		store:  storeBucket,
		logger: logger,
	}, nil
}

func (c *Client) initDefaultAdminPolicy() error {
	if c.cfg.DisableDefaultAdminPolicy {
		return nil
	}
	level.Info(c.logger).Log("msg", "init default admin policy")
	policy, err := c.GetAccessPolicy(context.Background(), DefaultAdminPolicyName)
	if err != nil {
		return err
	}
	if policy == nil {
		level.Info(c.logger).Log("msg", "start to create default admin policy")
		policy = &store.AccessPolicy{
			Name:        DefaultAdminPolicyName,
			DisplayName: "Admin Policy",
			Realms: []*store.Realm{
				&store.Realm{
					Tenant:  "*",
					Cluster: "",
				},
			},
			Scopes: []string{
				store.ADMIN,
			},
		}
		err := c.CreateAccessPolicy(context.Background(), policy)
		if err != nil {
			level.Error(c.logger).Log("msg", "failed to create default policy")
			return err
		}
	}
	return nil
}

func (c *Client) ListClusters(ctx context.Context) (*store.Clusters, error) {
	return c.store.ListClusters(ctx)
}

func (c *Client) GetCluster(ctx context.Context, name, kind string) (*store.Cluster, error) {
	return c.store.GetCluster(ctx, name, kind)
}

func (c *Client) ListTenants(ctx context.Context, includeNonActive bool) (*store.Tenants, error) {
	return c.store.ListTenants(ctx, includeNonActive)
}

func (c *Client) CreateTenant(ctx context.Context, tenant *store.Tenant) error {
	return c.store.CreateTenant(ctx, tenant)
}

func (c *Client) UpdateTenant(ctx context.Context, name string, tenant *store.Tenant) (*store.Tenant, error) {
	return c.store.UpdateTenant(ctx, name, tenant)
}

func (c *Client) GetTenant(ctx context.Context, name string) (*store.Tenant, error) {
	return c.store.GetTenant(ctx, name)
}

func (c *Client) ListAccessPolicies(ctx context.Context, includeNonActive bool) (*store.AccessPolicies, error) {
	return c.store.ListAccessPolicies(ctx, includeNonActive)
}

func (c *Client) CreateAccessPolicy(ctx context.Context, policy *store.AccessPolicy) error {
	return c.store.CreateAccessPolicy(ctx, policy)
}

func (c *Client) UpdateAccessPolicy(ctx context.Context, name string, policy *store.AccessPolicy) (*store.AccessPolicy, error) {
	return c.store.UpdateAccessPolicy(ctx, name, policy)
}

func (c *Client) GetAccessPolicy(ctx context.Context, name string) (*store.AccessPolicy, error) {
	return c.store.GetAccessPolicy(ctx, name)
}

func (c *Client) ListTokens(ctx context.Context, includeNonActive bool) (*store.Tokens, error) {
	return c.store.ListTokens(ctx, includeNonActive)
}

func (c *Client) CreateToken(ctx context.Context, token *store.Token) error {
	return c.store.CreateToken(ctx, token)
}

func (c *Client) GetToken(ctx context.Context, name string) (*store.Token, error) {
	return c.store.GetToken(ctx, name)
}

func (c *Client) DeleteToken(ctx context.Context, name string, token *store.Token) (*store.Token, error) {
	return c.store.UpdateToken(ctx, name, token)
}
