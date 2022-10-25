package api

func (a *API) RegisterComponent(component string, pageLinks ...IndexPageLink) {
	a.indexPage.AddLinks(defaultWeight, component, pageLinks)
}
