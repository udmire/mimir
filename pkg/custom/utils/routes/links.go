package routes

import (
	"github.com/grafana/mimir/pkg/api"
)

type Links interface {
	AddLink(alias, path string)
	AddDangerousLink(alias, path string)
	Register(group string, process func(group string, links ...api.IndexPageLink))
}

type internalLinks struct {
	links []api.IndexPageLink
}

func (i *internalLinks) Register(group string, process func(group string, links ...api.IndexPageLink)) {
	process(group, i.links...)
}

func (i *internalLinks) AddDangerousLink(alias, path string) {
	i.links = append(i.links, api.IndexPageLink{
		Desc:      alias,
		Path:      path,
		Dangerous: true,
	})
}

func (i *internalLinks) AddLink(alias, path string) {
	i.links = append(i.links, api.IndexPageLink{
		Desc:      alias,
		Path:      path,
		Dangerous: false,
	})
}
