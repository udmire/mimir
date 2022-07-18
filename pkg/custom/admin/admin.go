package admin

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/ring"
	"github.com/grafana/dskit/services"
	"github.com/grafana/mimir/pkg/custom/admin/store"
	"github.com/grafana/mimir/pkg/custom/utils/access"
	"github.com/grafana/mimir/pkg/custom/utils/token"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// RingKey is the key under which we store the admin ring in the KVStore.
	RingKey                           = "ring"
	instanceIngestionRateTickInterval = time.Second
)

type API struct {
	services.Service

	cfg ApiConfig

	client *Client
	signer token.TokenSigner

	ring       *ring.Ring
	lifecycler *ring.BasicLifecycler

	subservices        *services.Manager
	subservicesWatcher *services.FailureWatcher

	registry prometheus.Registerer
	logger   log.Logger
}

func NewAPI(cfg ApiConfig, client *Client, signer token.TokenSigner, log log.Logger, reg prometheus.Registerer) (*API, error) {
	subservices := []services.Service(nil)
	var err error

	if cfg.LeaderElection.Enabled {
		var leLifeCycler *ring.Lifecycler
		var leRing *ring.Ring

		leLifeCycler, err = ring.NewLifecycler(cfg.LeaderElection.Ring.ToLifecyclerConfig(), nil, "admin-api", RingKey, true, log, prometheus.WrapRegistererWithPrefix("cortex_", reg))
		if err != nil {
			return nil, err
		}

		leRing, err = ring.New(cfg.LeaderElection.Ring.ToRingConfig(), "admin-api", RingKey, log, prometheus.WrapRegistererWithPrefix("cortex_", reg))
		if err != nil {
			return nil, errors.Wrap(err, "failed to initialize distributors' ring client")
		}
		subservices = append(subservices, leLifeCycler, leRing)
	}

	a := &API{
		cfg:      cfg,
		client:   client,
		registry: reg,
		logger:   log,
		signer:   signer,
	}

	a.subservices, err = services.NewManager(subservices...)
	if err != nil {
		return nil, err
	}
	a.subservicesWatcher = services.NewFailureWatcher()
	a.subservicesWatcher.WatchManager(a.subservices)

	a.Service = services.NewBasicService(a.starting, a.run, a.stopping)
	return a, nil
}

func (a *API) starting(ctx context.Context) (err error) {
	if err = services.StartManagerAndAwaitHealthy(ctx, a.subservices); err != nil {
		return errors.Wrap(err, "unable to start admin-api subservices")
	}

	if a.lifecycler != nil && a.ring != nil {
		level.Info(a.logger).Log("msg", "waiting until api is ACTIVE in the ring")
		if err = ring.WaitInstanceState(ctx, a.ring, a.lifecycler.GetInstanceID(), ring.ACTIVE); err != nil {
			return err
		}
	}

	err = a.createBuiltinAdminPolicies()
	if err != nil {
		return err
	}

	return nil
}

func (a *API) run(ctx context.Context) error {
	ingestionRateTicker := time.NewTicker(instanceIngestionRateTickInterval)
	defer ingestionRateTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ingestionRateTicker.C:
			// a.ingestionRate.Tick()

		case err := <-a.subservicesWatcher.Chan():
			return errors.Wrap(err, "admin-api subservice failed")
		}
	}
}

func (a *API) stopping(_ error) error {
	if a.subservices != nil {
		return services.StopManagerAndAwaitStopped(context.Background(), a.subservices)
	}
	return nil
}

func (a *API) generateBuiltinAdminPolicy() *store.AccessPolicy {
	return &store.AccessPolicy{
		Name:        "__admin__",
		DisplayName: "Built-in Admin",
		Realms: []*store.Realm{
			{
				Tenant:  "*",
				Cluster: a.getClusterName(),
			},
		},
		Scopes: []string{
			access.ADMIN,
		},
	}
}

func (a *API) getClusterName() string {
	return a.cfg.ClusterName
}

func (a *API) createBuiltinAdminPolicies() error {
	if a.client.cfg.DisableDefaultAdminPolicy {
		level.Info(a.logger).Log("msg", "ignore create built-in admin policy")
		return nil
	}

	level.Debug(a.logger).Log("msg", "start to create built-in admin policy")
	policy := a.generateBuiltinAdminPolicy()
	_, err := a.client.GetAccessPolicy(context.Background(), "__admin__")
	if err == nil {
		level.Debug(a.logger).Log("msg", "built-in admin policy already exists")
		return nil
	}

	if err == store.ErrPolicyNotFound {
		return a.client.CreateAccessPolicy(context.Background(), policy)
	} else {
		level.Warn(a.logger).Log("msg", "internal error found while creating built-in admin policy")
		return err
	}
}
