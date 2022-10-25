package store

import (
	"context"
	"time"

	"github.com/grafana/mimir/pkg/custom/utils"
	"github.com/grafana/mimir/pkg/util/validation"
	"github.com/pkg/errors"
	"k8s.io/utils/strings/slices"
)

type Status string

const (
	ACTIVE   Status = "active"
	INACTIVE        = "inactive"
	UNKNOWN         = "unknown"
)

var (
	ErrClusterNotFound = errors.New("cluster not exist")
	ErrTenantNotFound  = errors.New("tenant not exist")
	ErrPolicyNotFound  = errors.New("policy not exist")
	ErrTokenNotFound   = errors.New("token not exist")
)

type Tenant struct {
	Name        string            `json:"name,omitempty"`
	DisplayName string            `json:"display_name,omitempty"`
	CreatedAt   time.Time         `json:"created_at,omitempty"`
	Status      Status            `json:"status,omitempty"`
	Cluster     string            `json:"cluster"`
	Limits      validation.Limits `json:"limits,omitempty"`
}

type Tenants struct {
	Type  string    `json:"type"`
	Items []*Tenant `json:"items"`
}

type AccessPolicy struct {
	Name        string    `json:"name,omitempty" yaml:"name"`
	DisplayName string    `json:"display_name,omitempty" yaml:"displayName"`
	CreatedAt   time.Time `json:"created_at,omitempty" yaml:"createdAt"`
	Status      Status    `json:"status,omitempty"`
	Realms      []*Realm  `json:"realms" yaml:"realms"`
	Scopes      []string  `json:"scopes" yaml:"scopes"`
}

func (a *AccessPolicy) HasScope(scope string) bool {
	return slices.Contains(a.Scopes, scope)
}

func (a *AccessPolicy) GetTenants() []string {
	var tenants []string
	for _, realm := range a.Realms {
		tenants = append(tenants, realm.Tenant)
	}
	return tenants
}

type Realm struct {
	Tenant        string              `json:"tenant" yaml:"tenant"`
	Cluster       string              `json:"cluster" yaml:"cluster"`
	LabelPolicies utils.LabelSelector `json:"label_policies,omitempty" yaml:"labelPolicies"`
}

func (r *Realm) GetTenant() string {
	return r.Tenant
}

func (r *Realm) GetCluster() string {
	return r.Cluster
}

func (r *Realm) GetLabelPolicies() utils.LabelSelector {
	return r.LabelPolicies
}

type AccessPolicies struct {
	Type  string          `json:"type" yaml:"type"`
	Items []*AccessPolicy `json:"items,omitempty" yaml:"items"`
}

type Token struct {
	Name         string    `json:"name,omitempty" yaml:"name"`
	DisplayName  string    `json:"display_name,omitempty" yaml:"displayName"`
	CreatedAt    time.Time `json:"created_at,omitempty" yaml:"createdAt"`
	CreatedBy    string    `json:"created_by,omitempty" yaml:"createdBy"`
	Status       Status    `json:"status,omitempty" yaml:"status"`
	AccessPolicy string    `json:"access_policy,omitempty" yaml:"accessPolicy"`
	Expiration   time.Time `json:"expiration,omitempty" yaml:"expiration"`
	Token        string    `json:"token,omitempty" yaml:"token"`
}

type Tokens struct {
	Type  string   `json:"type" yaml:"type"`
	Items []*Token `json:"items,omitempty" yaml:"items"`
}

type Cluster struct {
	Name        string    `json:"name,omitempty" yaml:"name"`
	DisplayName string    `json:"display_name,omitempty" yaml:"displayName"`
	CreatedAt   time.Time `json:"created_at,omitempty" yaml:"createdAt"`
	Kind        string    `json:"kind" yaml:"kind"`
	BaseUrl     string    `json:"base_url,omitempty" yaml:"baseUrl"`
}

type Clusters struct {
	Type  string     `json:"type" yaml:"type"`
	Items []*Cluster `json:"items,omitempty" yaml:"items"`
}

type ApiStore interface {
	ListClusters(ctx context.Context) (*Clusters, error)
	GetCluster(ctx context.Context, name, kind string) (*Cluster, error)
	CreateCluster(ctx context.Context, cluster *Cluster) error
	DeleteCluster(ctx context.Context, name, kind string) (*Cluster, error)

	ListTenants(ctx context.Context, includeNonActive bool) (*Tenants, error)
	CreateTenant(ctx context.Context, tenant *Tenant) error
	UpdateTenant(ctx context.Context, name string, tenant *Tenant) (*Tenant, error)
	GetTenant(ctx context.Context, name string) (*Tenant, error)

	ListAccessPolicies(ctx context.Context, includeNonActive bool) (*AccessPolicies, error)
	CreateAccessPolicy(ctx context.Context, policy *AccessPolicy) error
	UpdateAccessPolicy(ctx context.Context, name string, policy *AccessPolicy) (*AccessPolicy, error)
	GetAccessPolicy(ctx context.Context, name string) (*AccessPolicy, error)

	ListTokens(ctx context.Context, includeNonActive bool) (*Tokens, error)
	CreateToken(ctx context.Context, token *Token) error
	GetToken(ctx context.Context, name string) (*Token, error)
	UpdateToken(ctx context.Context, name string, token *Token) (*Token, error)
}
