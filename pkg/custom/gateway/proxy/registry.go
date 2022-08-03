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

	registry.Register(AdminApi, "/admin/api/**", []string{http.MethodGet}, true, true, access.ADMIN_READ, access.ADMIN)
	registry.Register(AdminApi, "/admin/api/**", []string{http.MethodPost, http.MethodPut}, true, true, access.ADMIN_READ, access.ADMIN)
	// Register Admin Links
	registry.RegisterLink(AdminApi, "Status", "/memberlist")

	registry.RegisterAll(Instance, "/node/api/**", true, true, access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Instance, "/", true, true, access.ADMIN, access.ADMIN_READ)

	registry.RegisterAll(Instance, "/config", true, true, access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Instance, "/runtime_config", true, true, access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Instance, "/services", true, true, access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Instance, "/memberlist", true, true, access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Instance, "/ready", true, true, access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Instance, "/metrics", true, false, access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Instance, "/debug/**", true, false, access.ADMIN, access.ADMIN_READ)

	registry.Register(Distributor, "/api/v1/push", []string{http.MethodPost}, true, true, access.METRICS_WRITE)
	registry.Register(Distributor, "/api/prom/push", []string{http.MethodPost}, true, true, access.METRICS_WRITE)
	registry.RegisterAll(Distributor, "/distributor/ring", true, true, access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Distributor, "/distributor/all_user_stats", true, true, access.ADMIN, access.ADMIN_READ)
	registry.RegisterAll(Distributor, "/distributor/ha_tracker", true, true, access.ADMIN, access.ADMIN_READ)
	// Register Distributor Links
	registry.RegisterLink(Distributor, "Ring status", "/distributor/ring")
	registry.RegisterLink(Distributor, "Usage statistics", "/distributor/all_user_stats")
	registry.RegisterLink(Distributor, "HA tracker status", "/distributor/ha_tracker")

	registry.RegisterAll(Ingester, "/ingester/flush", true, true, access.ADMIN)
	registry.RegisterAll(Ingester, "/ingester/shutdown", true, true, access.ADMIN)
	registry.Register(Ingester, "/ingester/ring", []string{http.MethodGet, http.MethodPost}, true, true, access.ADMIN, access.ADMIN_READ)
	// Register ingester Links
	registry.RegisterDangerousLink(Ingester, "Trigger a flush of data from ingester to storage", "/ingester/flush")
	registry.RegisterDangerousLink(Ingester, "Trigger ingester shutdown", "/ingester/shutdown")
	registry.RegisterLink(Ingester, "Ring status", "/ingester/ring")

	registry.RegisterAll(QueryFrontend, "/{prometheus}/api/v1/query", true, true, access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{prometheus}/api/v1/query_range", true, true, access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{prometheus}/api/v1/query_exemplars", true, true, access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{prometheus}/api/v1/series", true, true, access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{prometheus}/api/v1/labels", true, true, access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{prometheus}/api/v1/label/{name}/values", true, true, access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{prometheus}/api/v1/metadata", true, true, access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{prometheus}/api/v1/read", true, true, access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{prometheus}/api/v1/cardinality/label_names", true, true, access.METRICS_READ)
	registry.RegisterAll(QueryFrontend, "/{prometheus}/api/v1/cardinality/label_values", true, true, access.METRICS_READ)
	registry.Register(QueryFrontend, "/{prometheus}/api/v1/status/buildinfo", []string{http.MethodGet}, true, true, access.METRICS_READ)

	registry.Register(Querier, "/api/v1/user_stats", []string{http.MethodGet}, true, true, access.METRICS_READ)

	registry.Register(StoreGateway, "/store-gateway/ring", []string{http.MethodGet}, true, true, access.ADMIN_READ, access.ADMIN)
	registry.Register(StoreGateway, "/store-gateway/tenants", []string{http.MethodGet}, true, true, access.ADMIN_READ, access.ADMIN)
	registry.Register(StoreGateway, "/store-gateway/tenant/{tenant}/blocks", []string{http.MethodGet}, true, true, access.ADMIN_READ, access.ADMIN)
	// Register Store-Gateway Links
	registry.RegisterLink(StoreGateway, "Ring status", "/store-gateway/ring")
	registry.RegisterLink(StoreGateway, "Tenants & Blocks", "/store-gateway/tenants")

	registry.RegisterAll(Ruler, "/ruler/ring", true, true, access.ADMIN, access.ADMIN_READ)
	registry.Register(Ruler, "/ruler/delete_tenant_config", []string{http.MethodPost}, true, true, access.ADMIN, access.ADMIN_READ)
	registry.Register(Ruler, "/ruler/rule_groups", []string{http.MethodGet}, true, true, access.ADMIN, access.ADMIN_READ)
	registry.Register(Ruler, "/{prometheus}/api/v1/rules", []string{http.MethodGet}, true, true, access.RULES_READ)
	registry.Register(Ruler, "/{prometheus}/api/v1/alerts", []string{http.MethodGet}, true, true, access.RULES_READ)
	registry.Register(Ruler, "/{prometheus}/config/v1/rules", []string{http.MethodGet}, true, true, access.RULES_READ)
	registry.Register(Ruler, "/{prometheus}/config/v1/rules/{namespace}", []string{http.MethodGet}, true, true, access.RULES_READ)
	registry.Register(Ruler, "/{prometheus}/config/v1/rules/{namespace}/{groupName}", []string{http.MethodGet}, true, true, access.RULES_READ)
	registry.Register(Ruler, "/{prometheus}/config/v1/rules/{namespace}", []string{http.MethodPost}, true, true, access.RULES_WRITE)
	registry.Register(Ruler, "/{prometheus}/config/v1/rules/{namespace}/{groupName}", []string{http.MethodDelete}, true, true, access.RULES_WRITE)
	registry.Register(Ruler, "/{prometheus}/config/v1/rules/{namespace}", []string{http.MethodDelete}, true, true, access.RULES_WRITE)

	// Deduplicate
	registry.Register(Ruler, "/api/v1/rules", []string{http.MethodGet}, true, true, access.RULES_READ)
	registry.Register(Ruler, "/api/v1/rules/{namespace}", []string{http.MethodGet}, true, true, access.RULES_READ)
	registry.Register(Ruler, "/api/v1/rules/{namespace}/{groupName}", []string{http.MethodGet}, true, true, access.RULES_READ)
	registry.Register(Ruler, "/api/v1/rules/{namespace}", []string{http.MethodPost}, true, true, access.RULES_WRITE)
	registry.Register(Ruler, "/api/v1/rules/{namespace}/{groupName}", []string{http.MethodDelete}, true, true, access.RULES_WRITE)
	registry.Register(Ruler, "/api/v1/rules/{namespace}", []string{http.MethodDelete}, true, true, access.RULES_WRITE)
	_ = registry.RegisterRewrite(Ruler, "/api/v1/rules**", []string{http.MethodGet, http.MethodPost, http.MethodDelete}, "/api/v1/rules(.*)", "/prometheus/config/v1/rules$1")
	// Register Ruler Links
	registry.RegisterLink(Ruler, "Ring status", "/ruler/ring")
	registry.RegisterLink(Ruler, "Rule Groups", "/ruler/rule_groups")

	registry.RegisterAll(Compactor, "/compactor/ring", true, true, access.ADMIN, access.ADMIN_READ)
	registry.Register(Compactor, "/api/v1/upload/block/{block}", []string{http.MethodPost}, true, true, access.ADMIN)
	registry.Register(Compactor, "/api/v1/upload/block/{block}/files", []string{http.MethodPost}, true, true, access.ADMIN)
	// Register Compactor Links
	registry.RegisterLink(Compactor, "Ring status", "/compactor/ring")

	registry.Register(AlertManager, "/multitenant_alertmanager/status", []string{http.MethodGet}, true, true, access.ADMIN_READ)
	registry.Register(AlertManager, "/multitenant_alertmanager/configs", []string{http.MethodGet}, true, true, access.ADMIN_READ)
	registry.RegisterAll(AlertManager, "/multitenant_alertmanager/ring", true, true, access.ADMIN_READ)
	registry.Register(AlertManager, "/multitenant_alertmanager/delete_tenant_config", []string{http.MethodPost}, true, true, access.ADMIN)
	registry.Register(AlertManager, "/alertmanager/api/v1/status/buildinfo", []string{http.MethodGet}, false, true)
	registry.Register(AlertManager, "/api/v1/alerts", []string{http.MethodGet}, true, true, access.ALERTS_READ)
	registry.Register(AlertManager, "/api/v1/alerts", []string{http.MethodPost, http.MethodDelete}, true, true, access.ALERTS_WRITE)
	registry.Register(AlertManager, "/alertmanager/**", []string{http.MethodPost}, true, true, access.ALERTS_WRITE)
	registry.Register(AlertManager, "/alertmanager/**", []string{http.MethodGet}, true, true, access.ALERTS_READ)
	// Register Alertmanager Links
	registry.RegisterLink(AlertManager, "Status", "/multitenant_alertmanager/status")
	registry.RegisterLink(AlertManager, "Ring status", "/multitenant_alertmanager/ring")
	registry.RegisterLink(AlertManager, "UI", "/alertmanager/")

	registry.Register(Purger, "/purger/delete_tenant_status", []string{http.MethodGet}, true, true, access.ADMIN_READ, access.ADMIN, access.METRICS_DELETE)
	registry.RegisterStrict(Purger, "/purger/delete_tenant", []string{http.MethodPost}, true, true, access.ADMIN, access.METRICS_DELETE)
}
