package proxy

import (
	"net/http"
)

type dynamicTarget func(instance string) string
type requestModifier func(req *http.Request)

// type dynamicHttpReverseProxyWrapper struct {
// 	proxy   ReverseProxy
// 	matcher *delegateMatcher
// }
//
// type delegateMatcher struct {
// 	rootPattern *utils.AntPattern
// 	delegate    utils.Matcher
// }
//
// func (d *delegateMatcher) Matches(path string) bool {
// 	if !d.rootPattern.Matches(path) {
// 		return false
// 	}
// 	_, delegatePath := d.getDelegatePath(path)
// 	return d.delegate.Matches(delegatePath)
// }
//
// func (d *delegateMatcher) getDelegatePath(path string) (string, string) {
// 	groups := d.rootPattern.FindStringSubmatch(path)
// 	if len(groups) < 2 {
// 		return groups[0], "/"
// 	}
// 	return groups[0], groups[1]
// }

// func (d *dynamicHttpReverseProxyWrapper) Get() ReverseProxy {
// 	return d.proxy
// }
//
// func (d *dynamicHttpReverseProxyWrapper) Matches(path string) bool {
// 	return d.matcher.Matches(path)
// }

// func newDynamicInstanceProxy(cfg InstanceProxyConfig, subMatcher *routes.ComponentRoutes) (ReverseProxyWrapper, error) {
// 	if !cfg.Enabled {
// 		return nil, nil
// 	}
//
// 	matcher := &delegateMatcher{
// 		rootPattern: utils.MustCompile(routes.DynamicInstanceRoute),
// 		delegate:    subMatcher.HttpRoutes,
// 	}
//
// 	target := func(instance string) string {
// 		return fmt.Sprintf(cfg.Pattern, instance)
// 	}
//
// 	modifier := func(req *http.Request) {
// 		_, path := matcher.getDelegatePath(req.RequestURI)
// 		req.URL.Path = path
// 	}
//
// 	proxy, err := NewDynamicProxy(target, modifier)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return &dynamicHttpReverseProxyWrapper{
// 		matcher: matcher,
// 		proxy:   proxy,
// 	}, nil
// }
