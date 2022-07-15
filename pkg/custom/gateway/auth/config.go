package auth

import (
	"flag"
	"time"

	"github.com/go-kit/log"
)

type OverrideConfig struct {
	Token     string `yaml:"token" category:"advanced"`
	TokenFile string `yaml:"token_file" category:"advanced"`
}

func (c *OverrideConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.Token, "auth.override.token", "", "Override admin token. If set, this string will always be accepted as a token with admin level scope.")
	f.StringVar(&c.TokenFile, "auth.override.token-file", "", "If set, this file will be read at startup and the string from that file will be used as a admin scoped token.")
}

type AdminConfig struct {
	CacheTTL time.Duration `yaml:"cache_ttl" category:"advanced"`
	OIDC     OidcConfig    `yaml:"oidc"`
	Hmac     HmacConfig    `yaml:"hmac,omitempty"`
}

func (c *AdminConfig) RegisterFlags(f *flag.FlagSet, logger log.Logger) {
	f.DurationVar(&c.CacheTTL, "auth.cache.ttl", 10*time.Minute, "how long auth responses should be cached.")
	c.OIDC.RegisterFlags(f, logger)
	c.Hmac.RegisterFlags(f)
}

type HmacConfig struct {
	Secret string `json:"secret"`
}

func (h *HmacConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&h.Secret, "auth.admin.hmac.secret", "", "Secret for hmac verifier and signer")
}

// Config for auth guardian usage.
type Config struct {
	Type               string         `yaml:"type"`
	RequiredForMetrics bool           `yaml:"required_for_metrics" category:"advanced"`
	Override           OverrideConfig `yaml:"override"`
	Admin              AdminConfig    `yaml:"admin"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet, logger log.Logger) {
	f.StringVar(&c.Type, "auth.type", "trust", "method for authenticating incoming HTTP requests, (trust, enterprise).")
	f.BoolVar(&c.RequiredForMetrics, "auth.required-for-metrics", false, "requires admin level auth for the /metrics endpoint.")
	c.Override.RegisterFlags(f)
	c.Admin.RegisterFlags(f, logger)
}

type OAuthClient struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type OidcConfig struct {
	IssuerUrl           string      `yaml:"issuer_url"`
	AccessPolicyClaim   string      `yaml:"access_policy_claim"`
	AccessPolicyRegex   string      `yaml:"access_policy_regex" category:"advanced"`
	Audience            string      `yaml:"audience"`
	DefaultAccessPolicy string      `yaml:"default_access_policy" category:"advanced"`
	AdfsCompatibility   bool        `yaml:"adfs_compatibility" category:"advanced"`
	Client              OAuthClient `yaml:"client,omitempty"`
	RedirectURL         string
	TokenDuration       time.Duration
}

func (c *OidcConfig) RegisterFlags(f *flag.FlagSet, logger log.Logger) {
	f.StringVar(&c.IssuerUrl, "auth.admin.oidc.issuer-url", "", "JWT token issuer URL (example \"https://accounts.google.com\")")
	f.StringVar(&c.AccessPolicyClaim, "auth.admin.oidc.access-policy-claim", "", "claim in the JWT token containing the access policy")
	f.StringVar(&c.AccessPolicyRegex, "auth.admin.oidc.access-policy-regex", "", "regex to extract the access policy from the JWT token. The first submatch of the provided regex expression will be used.")
	f.StringVar(&c.Audience, "auth.admin.oidc.audience", "", "optional audience to check in JWT token")
	f.StringVar(&c.DefaultAccessPolicy, "auth.admin.oidc.default-access-policy", "", "name of the access policy to use when the token doesn't contain an access policy")
	f.BoolVar(&c.AdfsCompatibility, "auth.admin.oidc.adfs-compatibility", false, "enable ADFS compatibility")
	c.Client.RegisterFlags(f, logger)
}

func (o *OAuthClient) RegisterFlags(f *flag.FlagSet, logger log.Logger) {
	f.StringVar(&o.ClientID, "auth.admin.oidc.client.id", "", "Client ID")
	f.StringVar(&o.ClientSecret, "auth.admin.oidc.client.secret", "", "Client secret")
}
