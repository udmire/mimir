package proxy

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/golang/snappy"
	"github.com/grafana/mimir/pkg/custom/gateway/auth"
	"github.com/grafana/mimir/pkg/mimirpb"
	"github.com/grafana/mimir/pkg/util"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
	"github.com/weaveworks/common/user"
)

const (
	DataPushRoute     = "/api/v1/push"
	TenantsDataHeader = "X-Scope-Tenants-Data"
	TenantsDataValue  = "true"
)

type tenantsPushProxy struct {
	cfg       *ComponentProxyConfig
	tenantCfg *TenantConfig

	cli *fasthttp.Client

	logger log.Logger
}

func NewTenantsPushProxy(cfg *ComponentProxyConfig, tenantCfg *TenantConfig, logger log.Logger) Proxy {
	return &tenantsPushProxy{
		cfg: cfg,
		cli: &fasthttp.Client{
			Name:               "gateway-tenants-push-client",
			ReadTimeout:        time.Duration(cfg.ReadTimeout),
			WriteTimeout:       time.Duration(cfg.WriteTimeout),
			MaxConnsPerHost:    128,
			MaxConnWaitTimeout: time.Second,
		},
		tenantCfg: tenantCfg,
		logger:    logger,
	}
}

type result struct {
	code int
	body []byte
	err  error
}

func (t *tenantsPushProxy) isTenantsData(req *http.Request) bool {
	if t.tenantCfg.MatchType != "header" {
		return false
	}

	value := req.Header.Get(TenantsDataHeader)
	return TenantsDataValue == value
}

func (t *tenantsPushProxy) Handler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		user := auth.GetPrincipal(req.Context())
		user.WrapRequest(req)

		code, body := 0, []byte("Ok")
		var errs *prometheus.MultiError
		var results []result

		if !t.isTenantsData(req) {
			var err error
			code, body, err = t.direct(req)
			if err != nil {
			}
			goto out
		} else {
			var r mimirpb.WriteRequest
			_, err := util.ParseProtoReader(req.Context(), req.Body, int(req.ContentLength), t.tenantCfg.MaxRecvMsgSize, nil, &r, util.RawSnappy)
			if err != nil {
				level.Error(t.logger).Log("err", err.Error())
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}
			results = t.dispatch(req, groupByUserID(r, t.tenantCfg.TenantLabel))
		}

		// Return 204 regardless of errors if AcceptAll is enabled
		if t.tenantCfg.AcceptAll {
			code, body = 204, nil
			goto out
		}

		for _, r := range results {
			if r.err != nil {
				errs.Append(r.err)
				level.Error(t.logger).Log("msg", "send tenant data error", "err", r.err)
				continue
			}

			if r.code < 200 || r.code >= 300 {
				level.Error(t.logger).Log("msg", string(r.body), "req_id", r.code)
			}

			if r.code > code {
				code, body = r.code, r.body
			}
		}

		if errs != nil && errs.MaybeUnwrap() != nil {
			http.Error(rw, errs.Error(), http.StatusInternalServerError)
			return
		}

	out:
		// Pass back max status code copyFrom upstream response
		rw.WriteHeader(code)
		rw.Write(body)
	})
}

func (t *tenantsPushProxy) RegisterRoutes(f func(path string, handler http.Handler, auth bool, gzipEnabled bool, method string, methods ...string)) {
	f(DataPushRoute, t.Handler(), true, true, http.MethodPost)
}

func (p *tenantsPushProxy) dispatch(req *http.Request, m map[string]*mimirpb.WriteRequest) (res []result) {
	var wg sync.WaitGroup
	res = make([]result, len(m))

	i := 0
	for tenant, wrReq := range m {
		wg.Add(1)
		go func(idx int, tenant string, wrReq *mimirpb.WriteRequest) {
			defer wg.Done()
			var r result
			r.code, r.body, r.err = p.send(req, tenant, wrReq)
			res[idx] = r
		}(i, tenant, wrReq)

		i++
	}

	wg.Wait()
	return
}

func (p *tenantsPushProxy) send(ori *http.Request, tenant string, wr *mimirpb.WriteRequest) (code int, body []byte, err error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	sp, _ := opentracing.StartSpanFromContext(ori.Context(), "Gateway.tenantProxy")
	sp.SetTag("tenant", tenant)
	defer sp.Finish()

	buf, err := p.marshal(wr)
	if err != nil {
		return
	}

	copyFrom(req, ori)
	if tenant != "" {
		req.Header.Set(user.OrgIDHeaderName, tenant)
	}

	req.SetBody(buf)

	req.SetRequestURI(p.cfg.Url + ori.RequestURI)

	if err = p.cli.Do(req, resp); err != nil {
		return
	}

	code = resp.Header.StatusCode()
	body = make([]byte, len(resp.Body()))
	copy(body, resp.Body())

	return
}

func (p *tenantsPushProxy) direct(ori *http.Request) (code int, body []byte, err error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	sp, _ := opentracing.StartSpanFromContext(ori.Context(), "Gateway.tenantProxy")
	defer sp.Finish()

	copyFrom(req, ori)

	req.SetBodyStream(ori.Body, req.Header.ContentLength())

	req.SetRequestURI(p.cfg.Url + ori.RequestURI)

	if err = p.cli.Do(req, resp); err != nil {
		return
	}

	code = resp.Header.StatusCode()
	body = make([]byte, len(resp.Body()))
	copy(body, resp.Body())

	return
}

func copyFrom(req *fasthttp.Request, ori *http.Request) {
	for name, values := range ori.Header {
		if len(values) > 0 {
			for _, value := range values {
				req.Header.Add(name, value)
			}
		}
	}
	req.Header.SetMethod(ori.Method)
}

func (p *tenantsPushProxy) marshal(wr *mimirpb.WriteRequest) (bufOut []byte, err error) {
	b := make([]byte, wr.Size())

	// Marshal to Protobuf
	if _, err = wr.MarshalTo(b); err != nil {
		return
	}

	// Compress with Snappy
	return snappy.Encode(nil, b), nil
}

func (t *tenantsPushProxy) Path() string {
	return DataPushRoute
}

func groupByUserID(req mimirpb.WriteRequest, tenantLabel string) map[string]*mimirpb.WriteRequest {
	// userID -> Series Map
	result := make(map[string]*mimirpb.WriteRequest)

	if len(req.Timeseries) == 0 {
		result[""] = &mimirpb.WriteRequest{
			Source:   req.Source,
			Metadata: req.Metadata,
		}
		return result
	}

	for _, series := range req.Timeseries {
		tenant := getValueAndRemoveLabel(tenantLabel, &series.Labels)

		userReq, ok := result[tenant]
		if !ok {
			userReq = &mimirpb.WriteRequest{
				Timeseries: []mimirpb.PreallocTimeseries{},
				Source:     req.Source,
				Metadata:   req.Metadata,
			}
			result[tenant] = userReq
		}
		userReq.Timeseries = append(userReq.Timeseries, series)
	}

	return result
}

func getValueAndRemoveLabel(labelName string, labels *[]mimirpb.LabelAdapter) string {
	for i := 0; i < len(*labels); i++ {
		pair := (*labels)[i]
		if pair.Name == labelName {
			*labels = append((*labels)[:i], (*labels)[i+1:]...)
			return pair.Value
		}
	}

	return ""
}
