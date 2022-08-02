package routes

type Registry interface {
	Register(group, pattern string, methods []string, permissions []string)
	RegisterStrict(group, pattern string, methods []string, permissions []string)
	RegisterAll(group, pattern string, permissions ...string)
	RegisterRewrite(group, pattern string, methods []string, regex, replacement string) error

	GetGroupRoutes(group string) []Route
	GetGroupRewrites(group string) []Rewriter
	GetGroupRoutesPermissions(group string) []RoutePermissions
	GetAllRoutesPermissions() []RoutePermissions
}

func NewRegistry() Registry {
	return &internalRegistry{
		groupedRoutes:   map[string][]RoutePermissions{},
		groupedRewrites: map[string][]Rewriter{},
	}
}

type internalRegistry struct {
	groupedRoutes   map[string][]RoutePermissions
	groupedRewrites map[string][]Rewriter
}

func (i *internalRegistry) Register(group, pattern string, methods []string, permissions []string) {
	route := WithPatternMethods(pattern, methods...)
	i.groupedRoutes[group] = append(i.groupedRoutes[group], With(route, permissions...))
}

func (i *internalRegistry) RegisterAll(group, pattern string, permissions ...string) {
	route := WithPattern(pattern)
	i.groupedRoutes[group] = append(i.groupedRoutes[group], With(route, permissions...))
}

func (i *internalRegistry) RegisterStrict(group, pattern string, methods []string, permissions []string) {
	route := WithPatternMethods(pattern, methods...)
	i.groupedRoutes[group] = append(i.groupedRoutes[group], StrictWith(route, permissions...))
}

func (i *internalRegistry) RegisterRewrite(group, pattern string, methods []string, regex, replacement string) error {
	route := WithPatternMethods(pattern, methods...)
	rewriter, err := NewRewriter(route, regex, replacement)
	if err != nil {
		return err
	}
	i.groupedRewrites[group] = append(i.groupedRewrites[group], rewriter)
	return nil
}

func (i *internalRegistry) GetGroupRoutes(group string) []Route {
	permissions, exists := i.groupedRoutes[group]
	if !exists {
		return []Route{}
	}

	var routes []Route
	for _, permission := range permissions {
		routes = append(routes, permission.GetRoute())
	}
	return routes
}

func (i *internalRegistry) GetGroupRewrites(group string) []Rewriter {
	rewrites, exists := i.groupedRewrites[group]
	if !exists {
		return []Rewriter{}
	}
	return rewrites
}

func (i *internalRegistry) GetGroupRoutesPermissions(group string) []RoutePermissions {
	permissions, exists := i.groupedRoutes[group]
	if !exists {
		return []RoutePermissions{}
	}
	return permissions
}

func (i *internalRegistry) GetAllRoutesPermissions() []RoutePermissions {
	var all []RoutePermissions
	for _, permissions := range i.groupedRoutes {
		all = append(all, permissions...)
	}
	return all
}
