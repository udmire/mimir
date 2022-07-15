package proxy

import (
	"net/http"
	"net/http/httputil"

	"github.com/go-kit/log"
	"github.com/grafana/mimir/pkg/custom/utils"
	_ "github.com/grafana/mimir/pkg/custom/utils/routes"
)

type ReverseProxy interface {
	Proxy(logger log.Logger, rw http.ResponseWriter, request *http.Request)
}

type reverseProxy struct {
	modifier requestModifier
	proxy    func(req *http.Request) (*httputil.ReverseProxy, error)
}

func (r *reverseProxy) Proxy(logger log.Logger, rw http.ResponseWriter, request *http.Request) {
	r.modifier(request)
	proxy, err := r.proxy(request)
	if err != nil {
		utils.JSONError(logger, rw, "", http.StatusInternalServerError)
		return
	}
	proxy.ServeHTTP(rw, request)
}

func NewHttpReverseProxy(modifier requestModifier, proxy func(req *http.Request) (*httputil.ReverseProxy, error)) ReverseProxy {
	return &reverseProxy{
		proxy:    proxy,
		modifier: modifier,
	}
}

type ReverseProxyWrapper interface {
	Get() ReverseProxy
	Matches(path string) bool
}

// type httpReverseProxyWrapper struct {
// 	proxy    ReverseProxy
// 	matchers utils.Matcher
// }

// func (h *httpReverseProxyWrapper) Get() ReverseProxy {
// 	return h.proxy
// }
//
// func (h *httpReverseProxyWrapper) Matches(path string) bool {
// 	return h.matchers.Matches(path)
// }

// func newReverseProxyWrapper(config *ComponentProxyConfig, route *routes.ComponentRoutes) (ReverseProxyWrapper, error) {
// 	proxy, err := NewProxy(config)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &httpReverseProxyWrapper{
// 		proxy:    proxy,
// 		matchers: route.HttpRoutes,
// 	}, nil
// }

// func WithProxy(next http.Handler, factory ReverseProxyFactory, logger log.Logger) http.Handler {
// 	return http.HandlerFunc(func(rw http.ResponseWriter, request *http.Request) {
// 		// TODO
// 		requestLogger := log.With(logger, "request", request.Context().Value("trace"))
// 		uri := request.RequestURI
// 		proxy, err := factory.GetReverseProxy(uri)
// 		if err != nil {
// 			utils.JSONError(logger, rw, "Authentication required", http.StatusUnauthorized)
// 			return
// 		}
//
// 		if proxy == nil {
// 			next.ServeHTTP(rw, request)
// 			return
// 		}
//
// 		principal := auth.GetPrincipal(request.Context())
// 		if principal != nil {
// 			principal.WrapRequest(request)
// 		}
// 		proxy.Proxy(requestLogger, rw, request)
// 	})
// }
