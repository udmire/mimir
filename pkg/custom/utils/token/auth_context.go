package token

import (
	"github.com/grafana/mimir/pkg/custom/utils"
)

type IRealm interface {
	GetTenant() string
	GetCluster() string
	GetLabelPolicies() utils.LabelSelector
}

type Policy interface {
	HasScope(scope string) bool
	GetTenants() []string
}

type AuthContext interface {
	GetPolicy() Policy
}

func NewAuthContext(policy Policy) AuthContext {
	return &singlePolicyAuthContext{
		policy: policy,
	}
}

type AuthContextLoader interface {
	LoadContext(claims *CustomClaims) (AuthContext, error)
}

type singlePolicyAuthContext struct {
	policy Policy
}

func (s *singlePolicyAuthContext) GetPolicy() Policy {
	return s.policy
}
