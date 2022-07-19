package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MustCompiles(t *testing.T) {
	tests := map[string]struct {
		pattern  string
		expected string
	}{
		"1": {
			pattern:  "/**",
			expected: "^/.*$",
		},
		"2": {
			pattern:  "/*/**",
			expected: "^/[^/]*/.*$",
		},
		"3": {
			pattern:  "/{param}/**",
			expected: "^/[^/]*/.*$",
		},
		"4": {
			pattern:  "/*/*",
			expected: "^/[^/]*/[^/]*$",
		},
	}

	for testName, testData := range tests {
		t.Run(testName, func(t *testing.T) {
			result := MustCompile(testData.pattern)
			assert.Equal(t, testData.expected, result.regexPattern.String())
		})
	}
}
