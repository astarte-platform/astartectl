package utils

import (
	"encoding/json"

	"sigs.k8s.io/yaml"
)

// UnmarshalYAMLToJSON converts a valid Kubernetes YAML resource to its JSON representation
func UnmarshalYAMLToJSON(content []byte) (map[string]interface{}, error) {
	j2, err := yaml.YAMLToJSON(content)
	if err != nil {
		return nil, err
	}
	var jsonStruct map[string]interface{}
	err = json.Unmarshal(j2, &jsonStruct)
	if err != nil {
		return nil, err
	}
	return jsonStruct, nil
}
