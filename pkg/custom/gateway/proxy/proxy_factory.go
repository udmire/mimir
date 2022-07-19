package proxy

import (
	"github.com/gorilla/mux"
)

type ReverseProxyFactory interface {
	Register(router *mux.Router)
}

func HasComponent(cc ComponentProxyConfig) bool {
	return cc.Url != ""
}
