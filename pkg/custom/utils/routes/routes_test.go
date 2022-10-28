package routes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ProcessPattern(t *testing.T) {
	tests := map[string]struct {
		pattern  string
		expected string
	}{
		"1": {
			pattern:  "/**",
			expected: "/{param0:.+}",
		},
		"2": {
			pattern:  "/*/**",
			expected: "/{param0}/{param1:.+}",
		},
		"3": {
			pattern:  "/{abc}/*/**",
			expected: "/{abc}/{param0}/{param1:.+}",
		},
		"4": {
			pattern:  "/abc/*/def/**",
			expected: "/abc/{param0}/def/{param1:.+}",
		},
		"5": {
			pattern:  "/**/abc/*/def/**",
			expected: "/{param0:.+}/abc/{param1}/def/{param2:.+}",
		},
	}
	for testName, testData := range tests {
		t.Run(testName, func(t *testing.T) {
			result := precessPattern(testData.pattern)
			assert.Equal(t, testData.expected, result)
		})
	}
}