package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-kit/log/level"
	"github.com/go-openapi/runtime/middleware/header"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/grafana/mimir/pkg/custom/admin/store"
	"github.com/grafana/mimir/pkg/custom/utils/access"
	"github.com/grafana/mimir/pkg/custom/utils/token"
	"github.com/grafana/mimir/pkg/util"
	util_log "github.com/grafana/mimir/pkg/util/log"
	"github.com/weaveworks/common/errors"
	"gopkg.in/yaml.v2"
)

func (a *API) ListClusters(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)
	level.Debug(logger).Log("msg", "retrieving clusters")

	cls, err := a.client.ListClusters(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(cls.Items) == 0 {
		level.Info(logger).Log("msg", "no clusters found")
	} else {
		level.Debug(logger).Log("msg", "retrieved clusters from store", "length", len(cls.Items))
	}

	util.WriteJSONResponse(w, cls)
}

func (a *API) GetCluster(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)

	name, err := PathVariable(req, "name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	level.Debug(logger).Log("msg", "retrieving cluster with name", "name", name)
	c, err := a.client.GetCluster(req.Context(), name, "cortex")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.WriteJSONResponse(w, c)
}

func (a *API) CreateCluster(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)
	level.Debug(logger).Log("msg", "retrieving clusters")

	var cluster store.Cluster
	err := ParseContent(req, &cluster)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	level.Debug(logger).Log("msg", "create cluster with name", cluster.Name)
	err = a.client.CreateCluster(req.Context(), &cluster)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.WriteJSONResponse(w, cluster)
}

func (a *API) ListTenants(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)

	level.Debug(logger).Log("msg", "retrieving tenants")

	includeNonActive, err := BoolParamVariable(req, "include-non-active", false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ts, err := a.client.ListTenants(req.Context(), includeNonActive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(ts.Items) == 0 {
		level.Info(logger).Log("msg", "no tenants found")
	} else {
		level.Debug(logger).Log("msg", "retrieved tenants from store", "length", len(ts.Items))
	}

	util.WriteJSONResponse(w, ts)
}
func (a *API) CreateTenant(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)

	tenant := &store.Tenant{}
	err := ParseContent(req, tenant)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	level.Debug(logger).Log("msg", "create tenant with name", tenant.Name)
	err = a.client.CreateTenant(req.Context(), tenant)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.WriteJSONResponse(w, tenant)
}

func (a *API) UpdateTenant(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)

	name, err := PathVariable(req, "name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tenant := &store.Tenant{}
	err = ParseContent(req, tenant)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	level.Debug(logger).Log("msg", "update/delete tenant with name", "name", name, "content", tenant)
	tenant, err = a.client.UpdateTenant(req.Context(), name, tenant)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.WriteJSONResponse(w, tenant)
}
func (a *API) GetTenant(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)

	name, err := PathVariable(req, "name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	level.Debug(logger).Log("msg", "retrieving tenant with name", "name", name)
	c, err := a.client.GetTenant(req.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.WriteJSONResponse(w, c)
}

func (a *API) ListAccessPolicies(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)

	level.Debug(logger).Log("msg", "retrieving access policies")

	includeNonActive, err := BoolParamVariable(req, "include-non-active", false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	aps, err := a.client.ListAccessPolicies(req.Context(), includeNonActive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(aps.Items) == 0 {
		level.Info(logger).Log("msg", "no access policies found")
	} else {
		level.Debug(logger).Log("msg", "retrieved access policies from store", "length", len(aps.Items))
	}

	util.WriteJSONResponse(w, aps)
}
func (a *API) CreateAccessPolicy(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)

	policy := &store.AccessPolicy{}
	err := ParseContent(req, policy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	level.Debug(logger).Log("msg", "create policy with name", policy.Name)
	err = a.client.CreateAccessPolicy(req.Context(), policy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.WriteJSONResponse(w, policy)
}
func (a *API) UpdateAccessPolicy(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)

	name, err := PathVariable(req, "name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	policy := &store.AccessPolicy{}
	err = ParseContent(req, policy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	level.Debug(logger).Log("msg", "update/delete policy with name", "name", name, "content", policy)
	policy, err = a.client.UpdateAccessPolicy(req.Context(), name, policy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.WriteJSONResponse(w, policy)
}
func (a *API) GetAccessPolicy(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)

	name, err := PathVariable(req, "name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	level.Debug(logger).Log("msg", "retrieving access policy with name", "name", name)
	c, err := a.client.GetAccessPolicy(req.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.WriteJSONResponse(w, c)
}

func (a *API) ListToken(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)

	level.Debug(logger).Log("msg", "retrieving tokens")

	includeNonActive, err := BoolParamVariable(req, "include-non-active", false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ts, err := a.client.ListTokens(req.Context(), includeNonActive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(ts.Items) == 0 {
		level.Info(logger).Log("msg", "no tokens found")
	} else {
		level.Debug(logger).Log("msg", "retrieved tokens from store", "length", len(ts.Items))
	}

	util.WriteJSONResponse(w, ts)
}
func (a *API) CreateToken(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)

	t := &store.Token{}
	err := ParseContent(req, t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	policy, err := a.client.GetAccessPolicy(req.Context(), t.AccessPolicy)
	if err != nil {
		level.Error(logger).Log("msg", "invalid access policy", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	level.Debug(logger).Log("msg", "create token with name", t.Name)
	err = a.client.CreateToken(req.Context(), t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isAdmin := false
	for _, scope := range policy.Scopes {
		if scope == access.ADMIN {
			isAdmin = true
			continue
		}
	}

	claims := ToClaims(t, isAdmin)
	tokenString, err := a.signer.Sign(claims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSONResponse(w, tokenString)
}

func (a *API) GetToken(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)

	name, err := PathVariable(req, "name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	level.Debug(logger).Log("msg", "retrieving token with name", "name", name)
	c, err := a.client.GetToken(req.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.WriteJSONResponse(w, c)
}
func (a *API) DeleteToken(w http.ResponseWriter, req *http.Request) {
	logger := util_log.WithContext(req.Context(), a.logger)

	name, err := PathVariable(req, "name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token := &store.Token{}
	err = ParseContent(req, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	level.Debug(logger).Log("msg", "delete policy with name", "name", name, "content", token)
	token, err = a.client.UpdateToken(req.Context(), name, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.WriteJSONResponse(w, token)
}

func ToClaims(t *store.Token, hasAdminScope bool) *token.CustomClaims {
	return &token.CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   t.AccessPolicy,
			IssuedAt:  jwt.NewNumericDate(t.CreatedAt.UTC()),
			ExpiresAt: jwt.NewNumericDate(t.Expiration.UTC()),
		},
		Name:  t.DisplayName,
		Admin: hasAdminScope,
	}
}

func PathVariable(req *http.Request, name string) (string, error) {
	vars := mux.Vars(req)
	value, exists := vars[name]
	if exists {
		return value, nil
	}
	return "", errors.Error(fmt.Sprintf("path variable %s not exists", name))
}
func ParamVariable(req *http.Request, name string, defaultValue string) string {
	if req.URL.Query().Has(name) {
		return req.URL.Query().Get(name)
	} else {
		return defaultValue
	}
}
func BoolParamVariable(req *http.Request, name string, defaultValue bool) (bool, error) {
	if req.URL.Query().Has(name) {
		return strconv.ParseBool(req.URL.Query().Get(name))
	} else {
		return defaultValue, nil
	}
}
func ParseContent(req *http.Request, s interface{}) error {
	if req.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(req.Header, "Content-Type")
		if value == "application/json" {
			dec := json.NewDecoder(req.Body)
			dec.DisallowUnknownFields()
			return dec.Decode(s)
		} else if value == "application/yaml" {
			dec := yaml.NewDecoder(req.Body)
			dec.SetStrict(true)
			return dec.Decode(s)
		}
	}
	msg := "Content-Type header is not application/json or application/yaml"
	return errors.Error(msg)
}
