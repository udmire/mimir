package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/mimir/pkg/custom/gateway/auth"
	"github.com/grafana/mimir/pkg/custom/utils/proxy"
	_ "github.com/grafana/mimir/pkg/custom/utils/routes"
	"github.com/grafana/mimir/pkg/custom/utils/token"
	"github.com/opentracing/opentracing-go"
)

type ReverseProxy interface {
	Proxy(logger log.Logger, rw http.ResponseWriter, request *http.Request)
}

type TargetFunc func(req *http.Request) *url.URL
type PathFunc func(req *http.Request) string

type reverseProxy struct {
	ProxyLogging   bool
	Target         TargetFunc
	RawPath        PathFunc
	SendUserHeader bool
	logger         log.Logger
}

func (r *reverseProxy) Proxy(logger log.Logger, rw http.ResponseWriter, req *http.Request) {
	modifyResponse := func(resp *http.Response) error {
		if resp.StatusCode == 401 {
			// The data source rejected the request as unauthorized, convert to 400 (bad request)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read data source response body: %w", err)
			}
			_ = resp.Body.Close()

			level.Info(logger).Log("msg", "Authentication to data source failed", "body", string(body), "statusCode", resp.StatusCode)
			msg := "Authentication to data source failed"
			*resp = http.Response{
				StatusCode:    400,
				Status:        "Bad Request",
				Body:          ioutil.NopCloser(strings.NewReader(msg)),
				ContentLength: int64(len(msg)),
				Header:        http.Header{},
			}
		}
		return nil
	}

	sp, ctx := opentracing.StartSpanFromContext(req.Context(), "Gateway.doProxy")
	defer sp.Finish()

	reverseProxy := proxy.NewReverseProxy(logger, r.director, proxy.WithModifyResponse(modifyResponse))
	r.logRequest(req)

	reverseProxy.ServeHTTP(rw, req.Clone(ctx))
}

func (r *reverseProxy) director(req *http.Request) {
	targetUrl := r.Target(req)
	req.URL.Scheme = targetUrl.Scheme
	req.URL.Host = targetUrl.Host
	req.Host = targetUrl.Host

	req.URL.RawPath = r.RawPath(req)

	unescapedPath, err := url.PathUnescape(req.URL.RawPath)
	if err != nil {
		level.Error(r.logger).Log("msg", "Failed to unescape raw path", "rawPath", req.URL.RawPath, "error", err)
		return
	}

	req.URL.Path = unescapedPath

	applyUserHeader(r.SendUserHeader, req, auth.GetPrincipal(req.Context()))

	req.Header.Set("User-Agent", fmt.Sprintf("Gateway/proxy"))
}

func (r *reverseProxy) logRequest(req *http.Request) {
	if !r.ProxyLogging {
		return
	}

	var body string
	if req.Body != nil {
		buffer, err := ioutil.ReadAll(req.Body)
		if err == nil {
			req.Body = ioutil.NopCloser(bytes.NewBuffer(buffer))
			body = string(buffer)
		}
	}

	level.Info(r.logger).Log("msg", "Proxying incoming request", "uri", req.RequestURI, "method", req.Method, "body", body)
}

// Set the X-Grafana-User header if needed (and remove if not)
func applyUserHeader(sendUserHeader bool, req *http.Request, user token.IPrincipal) {
	req.Header.Del("X-Scope-User")
	if sendUserHeader {
		req.Header.Set("X-Scope-User", user.GetClaims().Name)
	}
	user.WrapRequest(req)
}

func NewHttpReverseProxy(logger log.Logger, targetFunc func(req *http.Request) *url.URL, rawPathFunc func(req *http.Request) string) ReverseProxy {
	return &reverseProxy{
		logger:         logger,
		SendUserHeader: false,
		ProxyLogging:   true,
		Target:         targetFunc,
		RawPath:        rawPathFunc,
	}
}
