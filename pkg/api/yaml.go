package api

import (
	"gopkg.in/yaml.v3"
	"net/http"
)

// YAMLv3 use "gopkg.in/yaml.v3" to marshal `yaml.Node{}` in RuleGroup properly
type YAMLv3 struct {
	Data interface{}
}

var yamlContentType = []string{"application/x-yaml; charset=utf-8"}

func (r YAMLv3) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	bytes, err := yaml.Marshal(r.Data)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	return err
}

// WriteContentType (YAML) writes YAML ContentType for response.
func (r YAMLv3) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, yamlContentType)
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}
