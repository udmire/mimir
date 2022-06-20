package mimir

import (
	"context"
	"flag"

	"github.com/go-kit/log"
	"github.com/grafana/dskit/modules"
	"github.com/grafana/dskit/services"
	"github.com/grafana/mimir/pkg/custom/admin"
	"github.com/grafana/mimir/pkg/custom/auth"
	"github.com/grafana/mimir/pkg/custom/gateway"
	"github.com/grafana/mimir/pkg/custom/license"
	"github.com/grafana/mimir/pkg/custom/tokengen"
	util_log "github.com/grafana/mimir/pkg/util/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Gateway     string = "gateway"
	AdminApi    string = "admin-api"
	AdminClient string = "admin-client"
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
}

func (t *Mimir) initGateway() (serv services.Service, err error) {
	t.Gateway, err = gateway.NewGateway(t.Cfg.Gateway, prometheus.DefaultRegisterer, util_log.Logger)
	if err != nil {
		return nil, err
	}
	return t.Gateway, nil
}

func (t *Mimir) initAdminAPI() (services.Service, error) {
	t.Cfg.AdminApi.LeaderElection.Ring.ListenPort = t.Cfg.Server.GRPCListenPort
	
	aa, err := admin.NewAPI(t.Cfg.AdminApi, t.AdminClient, util_log.Logger, prometheus.DefaultRegisterer)
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

func (t *Mimir) customModuleManager(mm *modules.Manager, deps map[string][]string) {
	mm.RegisterModule(Gateway, t.initGateway)
	mm.RegisterModule(AdminApi, t.initAdminAPI)
	mm.RegisterModule(AdminClient, t.initAdminClient, modules.UserInvisibleModule)

	deps[Gateway] = []string{Server, API}
	deps[AdminApi] = []string{Server, API, AdminClient}
	deps[AdminClient] = []string{}
}
