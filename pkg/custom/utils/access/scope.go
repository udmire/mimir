package access

type ScopeMatcher interface {
	// HasScopes check whether owns all the given scopes
	HasScopes(scopes ...string) bool

	// HasAnyScope check whether own any of the given scopes
	HasAnyScope(scopes ...string) bool
}
