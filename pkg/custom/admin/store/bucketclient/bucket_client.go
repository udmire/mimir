package bucketclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/mimir/pkg/custom/admin/store"
	"github.com/grafana/mimir/pkg/storage/bucket"
	"github.com/pkg/errors"
	"github.com/thanos-io/thanos/pkg/objstore"
	"gopkg.in/yaml.v2"
)

const (
	// The bucket prefix under which all tenants rule groups are stored.
	tenantsPrefix  = "tenants"
	policiesPrefix = "policies"
	tokensPrefix   = "tokens"
	clustersPrefix = "clusters"

	loadConcurrency = 10
)

var (
	errInvalidClusterKey = errors.New("invalid cluster object key")
	errEmptyClusterKind  = errors.New("empty cluster kind")
	errEmptyClusterName  = errors.New("empty cluster name")

	errEmptyUser = errors.New("empty user")
)

// BucketApiStore is used to support the ApiStore interface against an object storage backend. It is implemented
// using the Thanos objstore.Bucket interface
type BucketApiStore struct {
	clustersBucket objstore.Bucket
	tenantsBucket  objstore.Bucket
	policiesBucket objstore.Bucket
	tokensBucket   objstore.Bucket

	logger log.Logger
}

func NewApiStoreBucket(bkt objstore.Bucket, logger log.Logger) store.ApiStore {
	return &BucketApiStore{
		clustersBucket: bucket.NewPrefixedBucketClient(bkt, clustersPrefix),
		tenantsBucket:  bucket.NewPrefixedBucketClient(bkt, tenantsPrefix),
		policiesBucket: bucket.NewPrefixedBucketClient(bkt, policiesPrefix),
		tokensBucket:   bucket.NewPrefixedBucketClient(bkt, tokensPrefix),
		logger:         logger,
	}
}

func (b *BucketApiStore) ListClusters(ctx context.Context) (*store.Clusters, error) {
	clusters := &store.Clusters{
		Type:  "cluster",
		Items: []*store.Cluster{},
	}

	err := b.clustersBucket.Iter(ctx, "", func(key string) error {
		kind, name, err := parseClusterObjectKey(key)
		if err != nil {
			level.Warn(b.logger).Log("msg", "invalid cluster object key found while listing clusters", "key", key, "err", err)

			// Do not fail just because of a spurious item in the bucket.
			return nil
		}
		clusters.Items = append(clusters.Items, &store.Cluster{
			Name: name,
			Kind: kind,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list clusters in admin store bucket: %w", err)
	}

	return clusters, nil
}

func (b *BucketApiStore) GetCluster(ctx context.Context, name, kind string) (*store.Cluster, error) {
	cluster := &store.Cluster{}

	objectKey := getComposedObjectKey(kind, name)

	reader, err := b.clustersBucket.Get(ctx, objectKey)
	if b.clustersBucket.IsObjNotFoundErr(err) {
		level.Debug(b.logger).Log("msg", "cluster does not exist", "key", objectKey)
		return nil, store.ErrClusterNotFound
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to get cluster %s", objectKey)
	}
	defer func() { _ = reader.Close() }()

	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read cluster %s", objectKey)
	}

	err = yaml.Unmarshal(buf, cluster)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal cluster %s", objectKey)
	}

	return cluster, nil
}

func (b *BucketApiStore) CreateCluster(ctx context.Context, cluster *store.Cluster) error {
	objectKey := getComposedObjectKey(cluster.Kind, cluster.Name)

	exists, err := b.clustersBucket.Exists(ctx, objectKey)
	if err != nil {
		return errors.Wrapf(err, "failed to get cluster %s", objectKey)
	}
	if exists {
		return errors.New(fmt.Sprintf("cluster with name %s already exists.", cluster.Name))
	}

	cluster.CreatedAt = time.Now()
	data, err := yaml.Marshal(cluster)
	if err != nil {
		return err
	}

	return b.clustersBucket.Upload(ctx, objectKey, bytes.NewBuffer(data))
}

func (b *BucketApiStore) DeleteCluster(ctx context.Context, name, kind string) (*store.Cluster, error) {
	cluster, err := b.GetCluster(ctx, name, kind)
	if err != nil {
		return nil, err
	}
	objectKey := getComposedObjectKey(kind, name)
	err = b.clustersBucket.Delete(ctx, objectKey)
	return cluster, err
}

func (b *BucketApiStore) ListTenants(ctx context.Context, includeNonActive bool) (*store.Tenants, error) {
	tenants := &store.Tenants{
		Type:  "tenant",
		Items: []*store.Tenant{},
	}

	err := b.tenantsBucket.Iter(ctx, "", func(tenant string) error {
		tenants.Items = append(tenants.Items, &store.Tenant{
			Name: tenant,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list tenents in admin store bucket: %w", err)
	}

	// TODO: Load Content
	return tenants, nil
}

func (b *BucketApiStore) CreateTenant(ctx context.Context, tenant *store.Tenant) error {
	objectKey := getComposedObjectKey(tenant.Name)

	exists, err := b.tenantsBucket.Exists(ctx, objectKey)
	if err != nil {
		return errors.Wrapf(err, "failed to get tenant %s", objectKey)
	}
	if exists {
		return errors.New(fmt.Sprintf("tenant with name %s already exists.", tenant.Name))
	}

	tenant.CreatedAt = time.Now()
	tenant.Status = store.ACTIVE

	data, err := yaml.Marshal(tenant)
	if err != nil {
		return err
	}

	return b.tenantsBucket.Upload(ctx, objectKey, bytes.NewBuffer(data))
}

func (b *BucketApiStore) UpdateTenant(ctx context.Context, name string, tenant *store.Tenant) (*store.Tenant, error) {
	local, err := b.GetTenant(ctx, name)
	if err != nil {
		return nil, err
	}

	if tenant.Status == store.INACTIVE {
		level.Debug(b.logger).Log("msg", "delete tenant with name", "name", name)
		local.Status = store.INACTIVE
	} else {
		level.Debug(b.logger).Log("msg", "update tenant with content", "name", name, "content", tenant)
		local.DisplayName = tenant.DisplayName
		local.Limits = tenant.Limits
		local.Status = store.ACTIVE
	}

	objectKey := getComposedObjectKey(local.Name)
	data, err := yaml.Marshal(local)
	if err != nil {
		return nil, err
	}

	err = b.tenantsBucket.Upload(ctx, objectKey, bytes.NewBuffer(data))
	return local, err
}

func (b *BucketApiStore) GetTenant(ctx context.Context, name string) (*store.Tenant, error) {
	tenant := &store.Tenant{}
	objectKey := getComposedObjectKey(name)

	reader, err := b.tenantsBucket.Get(ctx, objectKey)
	if b.tenantsBucket.IsObjNotFoundErr(err) {
		level.Debug(b.logger).Log("msg", "tenant does not exist", "key", objectKey)
		return nil, store.ErrTenantNotFound
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to get tenant %s", name)
	}
	defer func() { _ = reader.Close() }()

	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read tenant %s", name)
	}

	err = yaml.Unmarshal(buf, tenant)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal tenant %s", name)
	}

	return tenant, nil
}

func (b *BucketApiStore) ListAccessPolicies(ctx context.Context, includeNonActive bool) (*store.AccessPolicies, error) {
	policies := &store.AccessPolicies{
		Type:  "policy",
		Items: []*store.AccessPolicy{},
	}

	err := b.policiesBucket.Iter(ctx, "", func(policy string) error {
		policies.Items = append(policies.Items, &store.AccessPolicy{
			Name: policy,
		})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("unable to list policies in admin store bucket: %w", err)
	}

	// TODO: Load Content & Filter
	return policies, nil
}

func (b *BucketApiStore) CreateAccessPolicy(ctx context.Context, policy *store.AccessPolicy) error {
	objectKey := getComposedObjectKey(policy.Name)

	exists, err := b.policiesBucket.Exists(ctx, objectKey)
	if err != nil {
		return errors.Wrapf(err, "failed to get policy %s", objectKey)
	}
	if exists {
		return errors.New(fmt.Sprintf("policy with name %s already exists.", policy.Name))
	}

	policy.CreatedAt = time.Now()
	policy.Status = store.ACTIVE

	data, err := yaml.Marshal(policy)
	if err != nil {
		return err
	}

	return b.policiesBucket.Upload(ctx, objectKey, bytes.NewBuffer(data))
}

func (b *BucketApiStore) UpdateAccessPolicy(ctx context.Context, name string, policy *store.AccessPolicy) (*store.AccessPolicy, error) {
	local, err := b.GetAccessPolicy(ctx, name)
	if err != nil {
		return nil, err
	}

	if policy.Status == store.INACTIVE {
		level.Debug(b.logger).Log("msg", "delete policy with name", "name", name)
		local.Status = store.INACTIVE
	} else {
		level.Debug(b.logger).Log("msg", "update policy with content", "name", name, "content", policy)
		local.DisplayName = policy.DisplayName
		local.Realms = policy.Realms
		local.Scopes = policy.Scopes
		local.Status = store.ACTIVE
	}

	objectKey := getComposedObjectKey(local.Name)
	data, err := yaml.Marshal(local)
	if err != nil {
		return nil, err
	}

	err = b.policiesBucket.Upload(ctx, objectKey, bytes.NewBuffer(data))
	return local, err
}

func (b *BucketApiStore) GetAccessPolicy(ctx context.Context, name string) (*store.AccessPolicy, error) {
	policy := &store.AccessPolicy{}

	objKey := getComposedObjectKey(name)
	reader, err := b.policiesBucket.Get(ctx, name)
	if b.tenantsBucket.IsObjNotFoundErr(err) {
		level.Debug(b.logger).Log("msg", "policy does not exist", "key", objKey)
		return nil, store.ErrPolicyNotFound
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to get policy %s", name)
	}
	defer func() { _ = reader.Close() }()

	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read policy %s", name)
	}

	err = yaml.Unmarshal(buf, policy)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal policy %s", objKey)
	}

	return policy, nil
}

func (b *BucketApiStore) ListTokens(ctx context.Context, includeNonActive bool) (*store.Tokens, error) {
	tokens := &store.Tokens{
		Type:  "token",
		Items: []*store.Token{},
	}

	err := b.tokensBucket.Iter(ctx, "", func(policy string) error {
		tokens.Items = append(tokens.Items, &store.Token{
			Name: policy,
		})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("unable to list tokens in admin store bucket: %w", err)
	}

	// TODO: Load Content & Filter
	return tokens, nil
}

func (b *BucketApiStore) CreateToken(ctx context.Context, token *store.Token) error {
	objectKey := getComposedObjectKey(token.Name)

	exists, err := b.tokensBucket.Exists(ctx, objectKey)
	if err != nil {
		return errors.Wrapf(err, "failed to get token %s", objectKey)
	}
	if exists {
		return errors.New(fmt.Sprintf("token with name %s already exists.", token.Name))
	}

	token.CreatedAt = time.Now()
	token.CreatedBy = ""
	token.Status = store.ACTIVE

	data, err := yaml.Marshal(token)
	if err != nil {
		return err
	}

	return b.tokensBucket.Upload(ctx, objectKey, bytes.NewBuffer(data))
}

func (b *BucketApiStore) GetToken(ctx context.Context, name string) (*store.Token, error) {
	token := &store.Token{}
	objectKey := getComposedObjectKey(name)

	reader, err := b.tokensBucket.Get(ctx, objectKey)
	if b.tenantsBucket.IsObjNotFoundErr(err) {
		level.Debug(b.logger).Log("msg", "token does not exist", "name", objectKey)
		return nil, store.ErrTokenNotFound
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to get token %s", objectKey)
	}
	defer func() { _ = reader.Close() }()

	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read token %s", objectKey)
	}

	err = yaml.Unmarshal(buf, token)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal token %s", objectKey)
	}

	token.Token = ""
	return token, nil
}

func (b *BucketApiStore) UpdateToken(ctx context.Context, name string, token *store.Token) (*store.Token, error) {
	local, err := b.GetToken(ctx, name)
	if err != nil {
		return nil, err
	}

	if token.Status == store.INACTIVE {
		level.Debug(b.logger).Log("msg", "delete token with name", "name", name)
		local.Status = store.INACTIVE
	} else {
		level.Debug(b.logger).Log("msg", "update token with content", "name", name, "content", token)
		local.DisplayName = token.DisplayName
		local.Expiration = token.Expiration
		local.AccessPolicy = token.AccessPolicy
		local.CreatedBy = token.CreatedBy
		local.CreatedAt = token.CreatedAt
		local.Status = store.ACTIVE
	}

	objectKey := getComposedObjectKey(local.Name)
	data, err := yaml.Marshal(local)
	if err != nil {
		return nil, err
	}

	err = b.tokensBucket.Upload(ctx, objectKey, bytes.NewBuffer(data))
	return local, err
}

func getComposedObjectKey(sections ...string) string {
	if len(sections) == 0 {
		return ""
	}
	builder := strings.Builder{}
	isFirst := true
	for _, section := range sections {
		if !isFirst {
			builder.WriteString(objstore.DirDelim)
			isFirst = false
		}
		builder.WriteString(base64.URLEncoding.EncodeToString([]byte(section)))
	}
	return builder.String()
}

// parseRuleGroupObjectKey parses a bucket object key in the format "<namespace>/<rules group>".
func parseClusterObjectKey(key string) (kind, name string, _ error) {
	parts := strings.Split(key, objstore.DirDelim)
	if len(parts) != 2 {
		return "", "", errInvalidClusterKey
	}

	decodedKind, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", "", err
	}

	if len(decodedKind) == 0 {
		return "", "", errEmptyClusterKind
	}

	decodedName, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", err
	}

	if len(decodedName) == 0 {
		return "", "", errEmptyClusterName
	}

	return string(decodedKind), string(decodedName), nil
}
