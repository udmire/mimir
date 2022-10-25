package tokengen

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/mimir/pkg/custom/admin"
	"github.com/grafana/mimir/pkg/custom/admin/store"
	"github.com/grafana/mimir/pkg/custom/utils/token"
)

type Generator struct {
	cfg    Config
	client *admin.Client
	signer token.TokenSigner
	logger log.Logger
}

func New(config Config, client *admin.Client, logger log.Logger) (*Generator, error) {
	return &Generator{
		cfg:    config,
		client: client,
		logger: logger,
	}, nil
}

func (g *Generator) Generate() error {
	token := g.buildToken()
	err := g.client.CreateToken(context.Background(), token)
	if err != nil {
		level.Error(g.logger).Log("msg", "token generate failed", "name", token.Name, "err", err)
		return err
	}
	level.Info(g.logger).Log("msg", "token generated with name", "name", token.Name)
	claims := admin.ToClaims(token, true)

	tokenString, err := g.signer.Sign(claims)
	if err != nil {
		level.Error(g.logger).Log("msg", "token generate failed", "err", err)
		return err
	}

	if g.cfg.TokenFile != "" {
		return WriteTokenToFile(tokenString, g.cfg.TokenFile)
	}
	level.Info(g.logger).Log("msg", "token generated", "token", claims)
	return nil
}

func WriteTokenToFile(token, file string) error {
	token = fmt.Sprintf("%s\n", token)
	return os.WriteFile(file, []byte(token), os.ModeAppend)
}

func (g *Generator) buildToken() *store.Token {
	name := fmt.Sprintf("token-%d", time.Now().UnixMilli())
	return &store.Token{
		AccessPolicy: g.cfg.AccessPolicy,
		Name:         name,
		DisplayName:  name,
	}
}
