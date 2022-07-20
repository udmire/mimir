package routes

import (
	"net/http"

	"github.com/grafana/mimir/pkg/custom/utils"
	"github.com/grafana/mimir/pkg/custom/utils/access"
	"github.com/grafana/mimir/pkg/custom/utils/token"
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
	Default       = "default"
	Instance      = "instance"
)

const (
	DynamicInstanceRoutePrefix = "/dynamic/*"
	// DynamicInstanceRoute       = "/dynamic/(*)(**)"
)

// AccessPermission defines permissions required for the given action
type AccessPermission interface {
	Matches(action string) bool
	HasPermission(principal token.IPrincipal) bool
}

type AccessPermissions []AccessPermission

func (a AccessPermissions) Matches(path string) bool {
	for _, permission := range a {
		if permission.Matches(path) {
			return true
		}
	}
	return false
}

type httpRequestAccessPermissions struct {
	pattern     *utils.AntPattern
	strict      bool
	permissions []string
}

func httpAccessPermission(method, uri string, permissions ...string) AccessPermission {
	return &httpRequestAccessPermissions{
		pattern:     utils.MustCompile(method + uri),
		permissions: permissions,
	}
}
func strictHttpAccessPermission(method, uri string, permissions ...string) AccessPermission {
	return &httpRequestAccessPermissions{
		pattern:     utils.MustCompile(method + uri),
		strict:      true,
		permissions: permissions,
	}
}

func (h *httpRequestAccessPermissions) Matches(action string) bool {
	return h.pattern.Matches(action)
}

func (h *httpRequestAccessPermissions) HasPermission(principal token.IPrincipal) bool {
	if h.strict {
		return principal.HasScopes(h.permissions...)
	} else {
		return principal.HasAnyScope(h.permissions...)
	}
}

var (
	componentRoutesMap = map[string]*ComponentRoutes{}

	AdminApiRoutes = []AccessPermission{
		httpAccessPermission(http.MethodGet, "/admin/api/**", access.ADMIN_READ, access.ADMIN),
		httpAccessPermission(http.MethodPost, "/admin/api/**", access.ADMIN),
		httpAccessPermission(http.MethodPut, "/admin/api/**", access.ADMIN),
	}
	NodeApiRoutes = []AccessPermission{
		httpAccessPermission("*", "/node/api/**", access.ADMIN, access.ADMIN_READ),
	}
	AllComponentRoutes = []AccessPermission{
		httpAccessPermission("*", "/", access.ADMIN, access.ADMIN_READ),
		httpAccessPermission("*", "/config", access.ADMIN, access.ADMIN_READ),
		httpAccessPermission("*", "/runtime_config", access.ADMIN, access.ADMIN_READ),
		httpAccessPermission("*", "/services", access.ADMIN, access.ADMIN_READ),
		httpAccessPermission("*", "/ready", access.ADMIN, access.ADMIN_READ),
		httpAccessPermission("*", "/metrics", access.ADMIN, access.ADMIN_READ),
		httpAccessPermission("*", "/debug/*", access.ADMIN, access.ADMIN_READ),
	}
	DistributorRoutes = []AccessPermission{
		httpAccessPermission(http.MethodPost, "/api/v1/push", access.METRICS_WRITE),
		httpAccessPermission(http.MethodPost, "/api/prom/push", access.METRICS_WRITE),
		httpAccessPermission("*", "/distributor/**", access.ADMIN, access.ADMIN_READ),
	}
	IngesterRoutes = []AccessPermission{
		httpAccessPermission("*", "/ingester/flush", access.ADMIN),
		httpAccessPermission("*", "/ingester/shutdown", access.ADMIN),
		httpAccessPermission(http.MethodGet, "/ingester/ring", access.ADMIN, access.ADMIN_READ),
	}
	QueryFrontendRoutes = []AccessPermission{
		httpAccessPermission("*", "/*/api/v1/query", access.METRICS_READ),
		httpAccessPermission("*", "/*/api/v1/query_range", access.METRICS_READ),
		httpAccessPermission("*", "/*/api/v1/query_exemplars", access.METRICS_READ),
		httpAccessPermission("*", "/*/api/v1/series", access.METRICS_READ),
		httpAccessPermission("*", "/*/api/v1/labels", access.METRICS_READ),
		httpAccessPermission("*", "/*/api/v1/label/*/values", access.METRICS_READ),
		httpAccessPermission("*", "/*/api/v1/metadata", access.METRICS_READ),
		httpAccessPermission("*", "/*/api/v1/read", access.METRICS_READ),
		httpAccessPermission("*", "/*/api/v1/cardinality/label_names", access.METRICS_READ),
		httpAccessPermission("*", "/*/api/v1/cardinality/label_values", access.METRICS_READ),
		httpAccessPermission(http.MethodGet, "/*/api/v1/status/buildinfo", access.METRICS_READ),
	}
	QuerierRoutes = []AccessPermission{
		httpAccessPermission(http.MethodGet, "/api/v1/user_stats", access.METRICS_READ),
	}
	StoreGatewayRoutes = []AccessPermission{
		httpAccessPermission(http.MethodGet, "/store-gateway/**", access.ADMIN_READ, access.ADMIN),
	}
	RulerRoutes = []AccessPermission{
		httpAccessPermission("*", "/ruler/*", access.ADMIN, access.ADMIN_READ),
		httpAccessPermission(http.MethodGet, "/*/api/v1/rules", access.RULES_READ),
		httpAccessPermission(http.MethodGet, "/*/api/v1/alerts", access.RULES_READ),
		httpAccessPermission(http.MethodGet, "/*/config/v1/rules**", access.RULES_READ),
		httpAccessPermission(http.MethodPost, "/*/config/v1/rules**", access.RULES_WRITE),
		httpAccessPermission(http.MethodDelete, "/*/config/v1/rules**", access.RULES_WRITE),
	}
	CompactorRoutes = []AccessPermission{
		httpAccessPermission(http.MethodGet, "/compactor/*", access.ADMIN, access.ADMIN_READ),
	}
	AlertManagerRoutes = []AccessPermission{
		httpAccessPermission(http.MethodGet, "/multitenant_alertmanager/*", access.ADMIN_READ),
		httpAccessPermission(http.MethodPost, "/multitenant_alertmanager/*", access.ADMIN_READ, access.ADMIN),
		httpAccessPermission(http.MethodGet, "/api/v1/alerts", access.ALERTS_READ),
		httpAccessPermission(http.MethodPost, "/api/v1/alerts", access.ALERTS_WRITE),
		httpAccessPermission(http.MethodDelete, "/api/v1/alerts", access.ALERTS_WRITE),
		httpAccessPermission(http.MethodPost, "/alertmanager/**", access.ALERTS_WRITE),
		httpAccessPermission(http.MethodGet, "/alertmanager/**", access.ALERTS_READ),
	}
	PurgerRoutes = []AccessPermission{
		httpAccessPermission(http.MethodGet, "/purger/*", access.ADMIN_READ, access.ADMIN),
		strictHttpAccessPermission(http.MethodPost, "/purger/*", access.ADMIN, access.METRICS_DELETE),
	}
	DefaultRoutes = []AccessPermission{
		httpAccessPermission("*", "/**"),
	}
)

func init() {
	componentRoutesMap[Ingester] = &ComponentRoutes{HttpRoutes: IngesterRoutes}
	componentRoutesMap[Distributor] = &ComponentRoutes{HttpRoutes: DistributorRoutes}
	componentRoutesMap[AdminApi] = &ComponentRoutes{HttpRoutes: AdminApiRoutes}
	componentRoutesMap[QueryFrontend] = &ComponentRoutes{HttpRoutes: QueryFrontendRoutes}
	componentRoutesMap[StoreGateway] = &ComponentRoutes{HttpRoutes: StoreGatewayRoutes}
	componentRoutesMap[Ruler] = &ComponentRoutes{HttpRoutes: RulerRoutes}
	componentRoutesMap[Querier] = &ComponentRoutes{HttpRoutes: QuerierRoutes}
	componentRoutesMap[Compactor] = &ComponentRoutes{HttpRoutes: CompactorRoutes}
	componentRoutesMap[AlertManager] = &ComponentRoutes{HttpRoutes: AlertManagerRoutes}
	componentRoutesMap[Purger] = &ComponentRoutes{HttpRoutes: PurgerRoutes}
	componentRoutesMap[Default] = &ComponentRoutes{HttpRoutes: DefaultRoutes}
	componentRoutesMap[Instance] = &ComponentRoutes{
		HttpRoutes: append(AllComponentRoutes, NodeApiRoutes...),
	}
}

type ComponentRoutes struct {
	HttpRoutes AccessPermissions
	GrpcRoutes []string
}

func GetComponentRoutes(component string) *ComponentRoutes {
	return componentRoutesMap[component]
}

// func GetAllRoutesPermissions() []AccessPermission {
// 	var allAccessPermissions AccessPermissions
// 	allAccessPermissions = append(allAccessPermissions, DistributorRoutes...)
// 	allAccessPermissions = append(allAccessPermissions, QueryFrontendRoutes...)
// 	allAccessPermissions = append(allAccessPermissions, RulerRoutes...)
//
// 	allAccessPermissions = append(allAccessPermissions, AlertManagerRoutes...)
// 	allAccessPermissions = append(allAccessPermissions, QuerierRoutes...)
// 	allAccessPermissions = append(allAccessPermissions, IngesterRoutes...)
//
// 	allAccessPermissions = append(allAccessPermissions, StoreGatewayRoutes...)
// 	allAccessPermissions = append(allAccessPermissions, CompactorRoutes...)
// 	allAccessPermissions = append(allAccessPermissions, PurgerRoutes...)
//
// 	allAccessPermissions = append(allAccessPermissions, AllComponentRoutes...)
// 	allAccessPermissions = append(allAccessPermissions, AdminApiRoutes...)
// 	allAccessPermissions = append(allAccessPermissions, NodeApiRoutes...)
// 	allAccessPermissions = append(allAccessPermissions, DefaultRoutes...)
//
// 	return allAccessPermissions
// }
