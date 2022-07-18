package mimir

import (
	"context"
	"flag"

	"github.com/go-kit/log"
	"github.com/grafana/dskit/modules"
	"github.com/grafana/dskit/services"
	"github.com/grafana/mimir/pkg/custom/admin"
	"github.com/grafana/mimir/pkg/custom/gateway"
	"github.com/grafana/mimir/pkg/custom/gateway/auth"
	"github.com/grafana/mimir/pkg/custom/license"
	"github.com/grafana/mimir/pkg/custom/tokengen"
	"github.com/grafana/mimir/pkg/custom/utils/token"
	util_log "github.com/grafana/mimir/pkg/util/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Gateway     string = "gateway"
	AdminApi    string = "admin-api"
	AdminClient string = "admin-client"
	TokenGen    string = "tokengen"
)

type CustomConfig struct {
	ClusterName string             `yaml:"cluster_name"`
	Gateway     gateway.Config     `yaml:"gateway"`
	AdminApi    admin.ApiConfig    `yaml:"admin_api"`
	AdminClient admin.ClientConfig `yaml:"admin_client"`
	Auth        auth.Config        `yaml:"auth"`
	License     license.Config     `yaml:"license"`
	Tokengen    tokengen.Config    `yaml:"tokengen"`
}

func (c *CustomConfig) RegisterFlags(f *flag.FlagSet, logger log.Logger) {
	f.StringVar(&c.ClusterName, "cluster-name", "", "Unique ID of this cluster. If undefined the name in the license is used.")
	c.Gateway.RegisterFlags(f, logger)
	c.AdminClient.RegisterFlags(f)
	c.AdminApi.RegisterFlags(f, logger)
	c.Auth.RegisterFlags(f, logger)
	c.License.RegisterFlags(f)
	c.Tokengen.RegisterFlags(f)
}

type CustomModule struct {
	Gateway     *gateway.Gateway
	AdminApi    *admin.API
	AdminClient *admin.Client
	TokenGen    *tokengen.Generator
}

func (t *Mimir) initGateway() (serv services.Service, err error) {
	logger := util_log.Logger

	t.Gateway, err = gateway.NewGateway(t.Cfg.Gateway, t.Cfg.Auth, t.AdminClient, prometheus.DefaultRegisterer, logger)
	if err != nil {
		return nil, err
	}

	t.Gateway.Register(t.Server.HTTP)

	return nil, nil
}

func (t *Mimir) initAdminAPI() (services.Service, error) {
	t.Cfg.AdminApi.LeaderElection.Ring.ListenPort = t.Cfg.Server.GRPCListenPort
	if t.Cfg.AdminApi.LeaderElection.Ring.KVStore.Store == "memberlist" {
		t.Cfg.AdminApi.LeaderElection.Ring.KVStore.MemberlistKV = t.MemberlistKV.GetMemberlistKV
	}

	signer := token.NewSigner([]byte(t.Cfg.Auth.Admin.Hmac.Secret))
	aa, err := admin.NewAPI(t.Cfg.AdminApi, t.AdminClient, signer, util_log.Logger, prometheus.DefaultRegisterer)
	if err != nil {
		return nil, err
	}

	t.AdminApi = aa
	t.AdminApi.RegisterAPI(t.API)
	return nil, nil
}

func (t *Mimir) initAdminClient() (serv services.Service, err error) {
	t.AdminClient, err = admin.NewClient(context.Background(), t.Cfg.AdminClient, util_log.Logger, prometheus.DefaultRegisterer)
	if err != nil {
		return nil, err
	}
	return
}

func (t *Mimir) initTokenGen() (serv services.Service, err error) {
	t.TokenGen, err = tokengen.New(t.Cfg.Tokengen, t.AdminClient, util_log.Logger)
	if err != nil {
		return nil, err
	}
	return
}

func (t *Mimir) customModuleManager(mm *modules.Manager, deps map[string][]string) {
	mm.RegisterModule(Gateway, t.initGateway)
	mm.RegisterModule(AdminApi, t.initAdminAPI)
	mm.RegisterModule(AdminClient, t.initAdminClient, modules.UserInvisibleModule)
	mm.RegisterModule(TokenGen, t.initTokenGen)

	deps[Gateway] = []string{Server, API, AdminClient}
	deps[AdminApi] = []string{Server, API, AdminClient, MemberlistKV}
	deps[AdminClient] = []string{}
	deps[TokenGen] = []string{AdminClient}
}
