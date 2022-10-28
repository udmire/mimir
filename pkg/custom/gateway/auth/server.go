package auth

import (
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/mimir/pkg/custom/admin"
	"github.com/grafana/mimir/pkg/custom/gateway/auth/access"
	"github.com/grafana/mimir/pkg/custom/utils/token"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	publicRoutes = []string{
		"/",
		"/config",
		"/runtime_config",
		"/services",
		"/debug/*",
	}
	metricRoute = "/metrics"
)

type AuthServer struct {
	cfg    Config
	logger log.Logger

	defaultToken string

	authChain PrincipalChain
	verifier  token.TokenVerifier
	loader    token.AuthContextLoader
	evaluator access.Evaluator

	authFailures *prometheus.CounterVec
	authSuccess  *prometheus.CounterVec
}

func NewAuthServer(cfg Config, eval access.Evaluator, client *admin.Client, logger log.Logger) (*AuthServer, error) {
	auth := &AuthServer{
		cfg:       cfg,
		logger:    logger,
		evaluator: eval,
	}

	auth.authFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cortex_gateway",
		Name:      "failed_authentications_total",
		Help:      "The total number of failed authentications.",
	}, []string{"reason"})
	auth.authSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cortex_gateway",
		Name:      "succeeded_authentications_total",
		Help:      "The total number of succeeded authentications.",
	}, []string{"tenant"})

	verifier, err := initVerifier(cfg, logger)
	if err != nil {
		return nil, err
	}
	auth.verifier = verifier

	auth.loader = NewAuthContextLoader(client, logger)

	err = auth.initOverrideToken()
	if err != nil {
		level.Warn(logger).Log("msg", err.Error())
	}
	auth.initAuthChain()

	return auth, nil
}

func initVerifier(cfg Config, logger log.Logger) (verifier token.TokenVerifier, err error) {
	switch cfg.Type {
	case "enterprise":
		verifier, err = NewOidcPrincipalVerifier(cfg.Admin.OIDC, logger)
		break
	case "trust":
		verifier = token.NewVerifier([]byte(cfg.Admin.Hmac.Secret))
	default:
		verifier = token.NewVerifier([]byte(cfg.Admin.Hmac.Secret))
		break
	}

	return
}

func (s *AuthServer) GetPublicRoutes() []string {
	var routes []string
	if !s.cfg.RequiredForMetrics {
		routes = append(routes, metricRoute)
	}
	routes = append(routes, publicRoutes...)
	return routes
}

func (s *AuthServer) OidcEnabled() bool {
	return s.cfg.Admin.OIDC.IssuerUrl != ""
}

func (s *AuthServer) HeaderEnabled() bool {
	return s.cfg.Admin.Header.Enabled && len(s.cfg.Admin.Header.DefaultTenants) > 0
}

func (s *AuthServer) initOverrideToken() error {
	if s.cfg.Override.Token != "" {
		s.defaultToken = s.cfg.Override.TokenFile
	}

	if s.cfg.Override.TokenFile != "" {
		content, err := os.ReadFile(s.cfg.Override.Token)
		if err != nil {
			level.Error(s.logger).Log("msg", "Read token from file failed", "err", err)
			return err
		}
		s.defaultToken = string(content)
	}
	return nil
}

func (s *AuthServer) initAuthChain() {
	basicAuth := BasicAuthPrincipalReader(s.logger, s.verifier)
	chain := PrincipalChain{basicAuth}

	if s.OidcEnabled() {
		bearerTokenAuth := BearerTokenPrincipalReader(s.logger, s.verifier)
		chain = append(chain, bearerTokenAuth)
	}

	if s.HeaderEnabled() {
		headerAuth := HeaderPrincipalReader(s.cfg.Admin.Header.HeaderName, s.cfg.Admin.Header.DefaultTenants...)
		chain = append(chain, headerAuth)
	}
	s.authChain = chain
}