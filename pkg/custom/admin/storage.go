// SPDX-License-Identifier: AGPL-3.0-only
// Provenance-includes-location: https://github.com/cortexproject/cortex/blob/master/pkg/ruler/storage.go
// Provenance-includes-license: Apache-2.0
// Provenance-includes-copyright: The Cortex Authors.

package admin

import (
	"context"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/mimir/pkg/custom/admin/store"
	"github.com/grafana/mimir/pkg/custom/admin/store/bucketclient"
	"github.com/grafana/mimir/pkg/storage/bucket"
	"github.com/prometheus/client_golang/prometheus"
)

// NewApiStoreBucket returns a api store backend client based on the provided cfg.
func NewApiStoreBucket(ctx context.Context, cfg StorageConfig, logger log.Logger, reg prometheus.Registerer) (store.ApiStore, error) {
	if cfg.Config.Backend == bucket.Filesystem {
		level.Warn(logger).Log("msg", "-admin.client.backend-type=filesystem is for development and testing only; you should switch to an external object store for production use or use a shared filesystem")
	}

	bucketClient, err := bucket.NewClient(ctx, cfg.Config, "admin-api-storage", logger, reg)
	if err != nil {
		return nil, err
	}

	store := bucketclient.NewApiStoreBucket(bucketClient, logger)
	if err != nil {
		return nil, err
	}

	return store, nil
}
