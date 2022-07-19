package utils

import (
	"bytes"
	"fmt"
	"regexp"
)

const (
	INITIAL = iota
	STAR_SEEN
	PARANTHESES_SEEN
)

type AntPattern struct {
	origStr      string
	specificity  int
	regexPattern *regexp.Regexp
}

func MustCompile(patternStr string) *AntPattern {
	pattern, err := ParseAntPattern(patternStr)
	if err != nil {
		panic(err)
	}

	return pattern
}

// ParseAntPattern
// /* => /[^/]*
// /** => /.*
// /*.jsp => /*\.jsp
func ParseAntPattern(patternStr string) (*AntPattern, error) {
	regexPatternStr, specificity, err := convertToRegex(patternStr)
	if err != nil {
		return nil, err
	}

	regexPattern, err := regexp.Compile(regexPatternStr)
	if err != nil {
		return nil, err
	}

	return &AntPattern{specificity: specificity, regexPattern: regexPattern, origStr: patternStr}, nil
}

func convertToRegex(antPatternStr string) (string, int, error) {
	var buf bytes.Buffer
	buf.WriteRune('^')

	state := INITIAL
	specificity := 0
	dotSpecifity := 0
	for _, chr := range antPatternStr {

		switch state {
		case INITIAL:
			if chr == '*' {
				state = STAR_SEEN
			} else if chr == '{' {
				state = PARANTHESES_SEEN
			} else if chr == '.' {
				buf.WriteString("\\.")
				dotSpecifity = 1
			} else if chr == '/' {
				specificity = specificity + 1
				buf.WriteRune(chr)
				dotSpecifity = 0
			} else {
				buf.WriteRune(chr)
			}
		case STAR_SEEN:
			if chr == '*' {
				buf.WriteString(".*")
				state = INITIAL
			} else if chr == '/' {
				buf.WriteString("[^/]*")
				specificity = specificity + 1
				buf.WriteRune(chr)
				dotSpecifity = 0
				state = INITIAL
			} else {
				buf.WriteString("[^/]*")
				buf.WriteRune(chr)
				state = INITIAL
			}
		case PARANTHESES_SEEN:
			if chr == '}' {
				state = INITIAL
				buf.WriteString("[^/]*")
			}
		default:
			return "", -1, fmt.Errorf("invalid state")
		}
	}

	buf.WriteRune('$')

	return buf.String(), specificity + dotSpecifity, nil
}

func (pattern *AntPattern) Specificity() int {
	return pattern.specificity
}

func (pattern *AntPattern) String() string {
	return pattern.origStr
}

func (pattern *AntPattern) Matches(path string) bool {
	return pattern.regexPattern.MatchString(path)
}

func (pattern *AntPattern) FindStringSubmatch(path string) []string {
	return pattern.regexPattern.FindStringSubmatch(path)
}
