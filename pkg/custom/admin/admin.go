package admin

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/dskit/ring"
	"github.com/grafana/dskit/services"
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

	ring       *ring.Ring
	lifecycler *ring.BasicLifecycler

	subservices        *services.Manager
	subservicesWatcher *services.FailureWatcher

	registry prometheus.Registerer
	logger   log.Logger
}

func NewAPI(cfg ApiConfig, client *Client, log log.Logger, reg prometheus.Registerer) (*API, error) {
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
	return services.StartManagerAndAwaitHealthy(ctx, a.subservices)
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
