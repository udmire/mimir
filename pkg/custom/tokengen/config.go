package tokengen

import (
	"flag"
)

type Config struct {
	AccessPolicy string `yaml:"access_policy" category:"advance"`
	TokenFile    string `yaml:"token_file" category:"advance"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.AccessPolicy, "tokengen.access-policy", "__admin__", "The name of the access policy to generate a token for. It defaults to the built-in admin policy.")
	f.StringVar(&c.TokenFile, "tokengen.token-file", "", "If set, the generated token will be written to a file at the provided path in addition to being logged.")
}
