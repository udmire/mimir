package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// LabelSelector define label selector silence
type LabelSelector []*labels.Matcher

// MarshalJSON implement json.Marshal
func (target *LabelSelector) MarshalJSON() ([]byte, error) {
	quoted := strconv.Quote(target.String())
	return []byte(quoted), nil
}

// UnmarshalJSON implement json.Unmarshal
func (target *LabelSelector) UnmarshalJSON(content []byte) error {
	unquoted, err := strconv.Unquote(string(content))
	if err != nil {
		return err
	}
	return target.fromString(unquoted)
}

// String implement stringer
func (target *LabelSelector) String() string {
	var strSlice []string
	if target != nil {
		for _, matcher := range *target {
			strSlice = append(strSlice, matcher.String())
		}
	}
	return fmt.Sprintf("{%s}", strings.Join(strSlice, ","))
}

func (target *LabelSelector) fromString(content string) error {
	selector, err := parser.ParseMetricSelector(content)
	if err != nil {
		return err
	}
	*target = selector
	return nil
}
