package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func Test_UnmarshalYaml(t *testing.T) {
	content := `default:
  url: http://10.97.99.157:8080
admin_api:
  url: http://mimir-read.infrastore.svc.cluster.local:8080
alertmanager:
  url: http://mimir-read.infrastore.svc.cluster.local:8080
`
	config := &Config{}
	err := yaml.Unmarshal([]byte(content), config)
	assert.Nilf(t, err, "No error")
}
