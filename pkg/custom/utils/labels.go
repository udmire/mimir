package utils

import (
	"bytes"
	"sort"
	"strconv"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// Labels define labels.Labels for refactor method
type Labels labels.Labels

// MarshalJSON implement json.Marshal
func (ls *Labels) MarshalJSON() ([]byte, error) {
	quoted := strconv.Quote(ls.String())
	return []byte(quoted), nil
}

// UnmarshalJSON implement json.Unmarshal
func (ls *Labels) UnmarshalJSON(content []byte) error {
	unquoted, err := strconv.Unquote(string(content))
	if err != nil {
		return err
	}
	return ls.FromString(unquoted)
}

// String implement stringer
func (ls *Labels) String() string {
	var b bytes.Buffer

	b.WriteByte('{')
	for i, l := range *ls {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(l.Name)
		b.WriteByte('=')
		b.WriteString(strconv.Quote(l.Value))
	}
	b.WriteByte('}')
	return b.String()
}

// FromString new labels from string
func (ls *Labels) FromString(content string) error {
	selector, err := parser.ParseMetric(content)
	if err != nil {
		return err
	}
	*ls = Labels(selector)

	return nil
}

// Copy deep cody labels
func (ls *Labels) Copy() *Labels {
	result := &Labels{}
	if ls != nil {
		for i := range *ls {
			label := (*ls)[i]
			*result = append(*result, labels.Label{Name: label.Name, Value: label.Value})
		}
	}
	return result
}

// WithoutLabels delete a label from labels
func (ls *Labels) WithoutLabels(names ...string) *Labels {
	ret := &Labels{}

	sort.Strings(names)
	j := 0
	if ls != nil {
		for i := range *ls {
			for j < len(names) && names[j] < (*ls)[i].Name {
				j++
			}
			if j < len(names) && (*ls)[i].Name == names[j] {
				continue
			}
			*ret = append(*ret, (*ls)[i])
		}
	}
	return ret
}
