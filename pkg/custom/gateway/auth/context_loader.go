package auth

import (
	"context"
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/mimir/pkg/custom/admin"
	"github.com/grafana/mimir/pkg/custom/admin/store"
	"github.com/grafana/mimir/pkg/custom/utils/token"
)

type remoteAuthContextLoader struct {
	logger log.Logger

	client *admin.Client
}

func NewAuthContextLoader(client *admin.Client, logger log.Logger) token.AuthContextLoader {
	return &remoteAuthContextLoader{
		logger: logger,
		client: client,
	}
}

func (a *remoteAuthContextLoader) LoadContext(claims *token.CustomClaims) (token.AuthContext, error) {
	if claims == nil {
		return nil, fmt.Errorf("claims must be provided")
	}

	var policy token.Policy
	accessPolicy, err := a.client.GetAccessPolicy(context.Background(), claims.GetPolicy())
	if err == nil {
		if accessPolicy.Status == store.ACTIVE {
			policy = accessPolicy
		} else {
			level.Debug(a.logger).Log("msg", "access policy is inactive", "name", accessPolicy)
		}
	} else if err == store.ErrPolicyNotFound {
		level.Warn(a.logger).Log("msg", "access policy is not exists")
		return nil, fmt.Errorf("access policy is not exists, maybe a misconfiguration for auth")
	} else {
		return nil, err
	}

	return token.NewAuthContext(policy), nil
}
