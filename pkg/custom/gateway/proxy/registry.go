package proxy

import (
	"net/http"

	"github.com/grafana/mimir/pkg/custom/utils/access"
	"github.com/grafana/mimir/pkg/custom/utils/routes"
)

const (
	Ingester      = "ingester"
	Distributor   = "distributor"
	AdminApi      = "admin-api"
	QueryFrontend = "query-frontend"
	StoreGateway  = "store-gateway"
	Ruler         = "ruler"
	Querier       = "querier"
	Compactor     = "compactor"
	AlertManager  = "alert-manager"
	Purger        = "purger"

	Default = "default"
)

func NewRegistry() routes.Registry {
	return routes.NewRegistry()
}

func Init(registry routes.Registry) {

	registry.Register(AdminApi, "/admin/api/**", []string{http.MethodGet}, []string{access.ADMIN_READ, access.ADMIN})
	registry.Register(AdminApi, "/admin/api/**", []string{http.MethodPost, http.MethodPut}, []string{access.ADMIN_READ, access.ADMIN})

	registry.RegisterAll(Instance, "/node/api/**", access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Instance, "/", access.ADMIN, access.ADMIN_READ)

	registry.RegisterAll(Instance, "/config", access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Instance, "/runtime_config", access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Instance, "/services", access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Instance, "/ready", access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Instance, "/metrics", access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Instance, "/debug/*", access.ADMIN, access.ADMIN_READ)

	registry.Register(Distributor, "/api/v1/push", []string{http.MethodPost}, []string{access.METRICS_WRITE})
	registry.Register(Distributor, "/api/prom/push", []string{http.MethodPost}, []string{access.METRICS_WRITE})
	registry.RegisterAll(Distributor, "/distributor/**", access.ADMIN, access.ADMIN_READ)

	registry.RegisterAll(Ingester, "/ingester/flush", access.ADMIN)
	registry.RegisterAll(Ingester, "/ingester/shutdown", access.ADMIN)
	registry.Register(Ingester, "/ingester/ring", []string{http.MethodGet}, []string{access.ADMIN, access.ADMIN_READ})

	registry.RegisterAll(QueryFrontend, "/{promPrefix}/api/v1/query", access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{promPrefix}/api/v1/query_range", access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{promPrefix}/api/v1/query_exemplars", access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{promPrefix}/api/v1/series", access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{promPrefix}/api/v1/labels", access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{promPrefix}/api/v1/label/{name}/values", access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{promPrefix}/api/v1/metadata", access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{promPrefix}/api/v1/read", access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{promPrefix}/api/v1/cardinality/label_names", access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{promPrefix}/api/v1/cardinality/label_values", access.METRICS_READ)
	registry.Register(QueryFrontend, "/{promPrefix}/api/v1/status/buildinfo", []string{http.MethodGet}, []string{access.METRICS_READ})

	registry.Register(Querier, "/api/v1/user_stats", []string{http.MethodGet}, []string{access.METRICS_READ})

	registry.Register(StoreGateway, "/store-gateway/**", []string{http.MethodGet}, []string{access.ADMIN_READ, access.ADMIN})

	registry.RegisterAll(Ruler, "/ruler/*", access.ADMIN, access.ADMIN_READ)
	registry.Register(Ruler, "/{promPrefix}/api/v1/rules", []string{http.MethodGet}, []string{access.RULES_READ})
	registry.Register(Ruler, "/{promPrefix}/api/v1/alerts", []string{http.MethodGet}, []string{access.RULES_READ})
	registry.Register(Ruler, "/{promPrefix}/config/v1/rules", []string{http.MethodGet}, []string{access.RULES_READ})
	registry.Register(Ruler, "/{promPrefix}/config/v1/rules", []string{http.MethodPost, http.MethodDelete},
		[]string{access.RULES_WRITE})

	registry.Register(Compactor, "/compactor/*", []string{http.MethodGet}, []string{access.ADMIN, access.ADMIN_READ})

	registry.Register(AlertManager, "/multitenant_alertmanager/*", []string{http.MethodGet}, []string{access.ADMIN_READ})
	registry.Register(AlertManager, "/multitenant_alertmanager/*", []string{http.MethodPost}, []string{access.ADMIN_READ, access.ADMIN})
	registry.Register(AlertManager, "/api/v1/alerts", []string{http.MethodGet}, []string{access.ALERTS_READ})
	registry.Register(AlertManager, "/api/v1/alerts", []string{http.MethodPost, http.MethodDelete}, []string{access.ALERTS_WRITE})
	registry.Register(AlertManager, "/alertmanager/**", []string{http.MethodPost}, []string{access.ALERTS_WRITE})
	registry.Register(AlertManager, "/alertmanager/**", []string{http.MethodGet}, []string{access.ALERTS_READ})

	registry.Register(Purger, "/purger/*", []string{http.MethodGet}, []string{access.ADMIN_READ, access.ADMIN})
	registry.RegisterStrict(Purger, "/purger/*", []string{http.MethodPost}, []string{access.ADMIN, access.METRICS_DELETE})

}
