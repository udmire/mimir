package license

import (
	"flag"
	"time"
)

type Config struct {
	Path         string        `yaml:"path"`
	SyncInterval time.Duration `yaml:"sync_interval" category:"advanced"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.Path, "license.path", "./license.jwt", "Filepath to license jwt file.")
	f.DurationVar(&c.SyncInterval, "license.sync-interval", time.Hour, "Interval to check for new or existing licenses.")
}
