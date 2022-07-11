package utils

type Matcher interface {
	Matches(path string) bool
}

type antMatchers struct {
	antPatterns []*AntPattern
}

func NewAntMatchers(patterns []string) Matcher {
	matcher := &antMatchers{
		antPatterns: []*AntPattern{},
	}
	if len(patterns) > 0 {
		for _, pattern := range patterns {
			matcher.antPatterns = append(matcher.antPatterns, MustCompile(pattern))
		}
	}

	return matcher
}

func (a *antMatchers) Matches(path string) bool {
	if len(a.antPatterns) == 0 {
		return false
	}

	for _, pattern := range a.antPatterns {
		if pattern.Matches(path) {
			return true
		}
	}
	return false
}
