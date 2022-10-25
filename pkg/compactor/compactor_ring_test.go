// SPDX-License-Identifier: AGPL-3.0-only
// Provenance-includes-location: https://github.com/cortexproject/cortex/blob/master/pkg/compactor/compactor_ring_test.go
// Provenance-includes-license: Apache-2.0
// Provenance-includes-copyright: The Cortex Authors.

package compactor

import (
	"fmt"
	"testing"
	"time"

	"github.com/grafana/dskit/flagext"
	"github.com/grafana/dskit/ring"
	"github.com/grafana/mimir/pkg/util/log"
	"github.com/stretchr/testify/assert"
)

func TestRingConfig_DefaultConfigToLifecyclerConfig(t *testing.T) {
	crc := RingConfig{}
	expectedRc := ring.Config{}
	flagext.DefaultValues(&crc, &expectedRc)

	expectedLc := ring.BasicLifecyclerConfig{}
	expectedLc.HeartbeatTimeout = crc.HeartbeatTimeout
	expectedLc.HeartbeatPeriod = crc.HeartbeatPeriod
	addr, _ := ring.GetInstanceAddr("", crc.InstanceInterfaceNames, log.Logger)
	port := ring.GetInstancePort(crc.InstancePort, crc.ListenPort)
	expectedLc.Addr = fmt.Sprintf("%s:%d", addr, port)
	lc, err := crc.ToLifecyclerConfig(log.Logger)
	rc := crc.ToRingConfig()

	// The default config of the compactor ring must be the exact same
	// of the default lifecycler config, except few options which are
	// intentionally overridden
	expectedLc.ID = lc.ID
	expectedRc.ReplicationFactor = 1
	expectedRc.SubringCacheDisabled = true
	expectedRc.KVStore.Store = "memberlist"
	expectedLc.NumTokens = 512
	expectedLc.HeartbeatPeriod = 15 * time.Second

	assert.Nil(t, err)
	assert.Equal(t, expectedLc, lc)
	assert.Equal(t, expectedRc, rc)
}

func TestRingConfig_CustomConfigToLifecyclerConfig(t *testing.T) {
	cfg := RingConfig{}
	flagext.DefaultValues(&cfg)

	// Customize the compactor ring config
	cfg.HeartbeatPeriod = 1 * time.Second
	cfg.HeartbeatTimeout = 10 * time.Second
	cfg.InstanceID = "test"
	cfg.InstanceInterfaceNames = []string{"abc1"}
	cfg.InstancePort = 10
	cfg.InstanceAddr = "1.2.3.4"
	cfg.ListenPort = 10

	expectedLc := ring.BasicLifecyclerConfig{}
	expectedRc := ring.Config{}
	flagext.DefaultValues(&expectedRc)
	// The lifecycler config should be generated based upon the compactor
	// ring config
	expectedLc.HeartbeatPeriod = cfg.HeartbeatPeriod
	expectedLc.HeartbeatTimeout = cfg.HeartbeatTimeout
	expectedRc.HeartbeatTimeout = cfg.HeartbeatTimeout
	expectedRc.SubringCacheDisabled = true
	expectedRc.KVStore.Store = "memberlist"
	expectedLc.ID = cfg.InstanceID
	expectedLc.Addr = fmt.Sprintf("%s:%d", cfg.InstanceAddr, cfg.InstancePort)

	// Hardcoded config
	expectedRc.ReplicationFactor = 1
	expectedLc.NumTokens = 512

	lc, _ := cfg.ToLifecyclerConfig(log.Logger)
	rc := cfg.ToRingConfig()
	assert.Equal(t, expectedLc, lc)
	assert.Equal(t, expectedRc, rc)
}
