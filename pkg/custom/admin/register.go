package admin

import (
	"net/http"
	"path"

	"github.com/grafana/mimir/pkg/api"
)

// RegisterAPI registers routes associated with the Admin API
func (a *API) RegisterAPI(api *api.API) {
	prefix := "/admin/api/v3"
	api.RegisterRoute(path.Join(prefix, "/clusters"), http.HandlerFunc(a.ListClusters), true, true, "GET")
	api.RegisterRoute(path.Join(prefix, "/clusters/{name}"), http.HandlerFunc(a.GetCluster), true, true, "GET")
	api.RegisterRoute(path.Join(prefix, "/clusters"), http.HandlerFunc(a.CreateCluster), true, true, "POST")

	api.RegisterRoute(path.Join(prefix, "/tenants"), http.HandlerFunc(a.ListTenants), true, true, "GET")
	api.RegisterRoute(path.Join(prefix, "/tenants"), http.HandlerFunc(a.CreateTenant), true, true, "POST")
	api.RegisterRoute(path.Join(prefix, "/tenants/{name}"), http.HandlerFunc(a.UpdateTenant), true, true, "PUT")
	api.RegisterRoute(path.Join(prefix, "/tenants/{name}"), http.HandlerFunc(a.GetTenant), true, true, "GET")

	api.RegisterRoute(path.Join(prefix, "/accesspolicies"), http.HandlerFunc(a.ListAccessPolicies), true, true, "GET")
	api.RegisterRoute(path.Join(prefix, "/accesspolicies"), http.HandlerFunc(a.CreateAccessPolicy), true, true, "POST")
	api.RegisterRoute(path.Join(prefix, "/accesspolicies/{name}"), http.HandlerFunc(a.UpdateAccessPolicy), true, true, "PUT")
	api.RegisterRoute(path.Join(prefix, "/accesspolicies/{name}"), http.HandlerFunc(a.GetAccessPolicy), true, true, "GET")

	api.RegisterRoute(path.Join(prefix, "/tokens"), http.HandlerFunc(a.ListToken), true, true, "GET")
	api.RegisterRoute(path.Join(prefix, "/tokens"), http.HandlerFunc(a.CreateToken), true, true, "POST")
	api.RegisterRoute(path.Join(prefix, "/tokens/{name}"), http.HandlerFunc(a.GetToken), true, true, "GET")
	api.RegisterRoute(path.Join(prefix, "/tokens/{name}"), http.HandlerFunc(a.DeleteToken), true, true, "PUT")
}
