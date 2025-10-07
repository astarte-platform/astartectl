package migrationutils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

// CopyIfExists copies a value from source object to destination object if the source key exists
// and the value is not nil. This prevents adding nil/zero values to the destination when
// a field is not present in the source or is explicitly set to nil/empty (e.g. by defaulting annotations).
func CopyIfExists(source *unstructured.Unstructured, dest *unstructured.Unstructured, sPath []string, dPath []string) error {
	if source == nil || dest == nil {
		return fmt.Errorf("source or dest is nil")
	}

	// Retrieve the field from source only if it exists
	sourceField, found, err := unstructured.NestedFieldNoCopy(source.Object, sPath...)
	if err != nil {
		return fmt.Errorf("error retrieving field %v from source: %w", sPath, err)
	}

	// If found and not nil/empty, set it in dest
	if found && sourceField != nil && sourceField != "" {
		if err := unstructured.SetNestedField(dest.Object, sourceField, dPath...); err != nil {
			return fmt.Errorf("error setting field %v in dest: %w", dPath, err)
		}
	}

	return nil
}

// SumResourceRequirements sums two v1.ResourceRequirements objects.
func SumResourceRequirements(a, b v1.ResourceRequirements) v1.ResourceRequirements {
	result := v1.ResourceRequirements{
		Limits:   sumResourceLists(a.Limits, b.Limits),
		Requests: sumResourceLists(a.Requests, b.Requests),
	}
	return result
}

// sumResourceLists sums two v1.ResourceList objects.
func sumResourceLists(rl1, rl2 v1.ResourceList) v1.ResourceList {
	if rl1 == nil && rl2 == nil {
		return nil
	}

	result := v1.ResourceList{}

	// Copy all from rl1
	for k, v := range rl1 {
		result[k] = v.DeepCopy()
	}

	// Add rl2 onto result
	for k, v := range rl2 {
		if existing, ok := result[k]; ok {
			sum := existing.DeepCopy()
			sum.Add(v)
			result[k] = sum
		} else {
			result[k] = v.DeepCopy()
		}
	}

	return result
}

// UnstructuredToResourceRequirements converts a generic interface (typically map[string]interface{})
// representing Kubernetes ResourceRequirements into a concrete v1.ResourceRequirements.
func UnstructuredToResourceRequirements(in interface{}) (v1.ResourceRequirements, error) {
	var rr v1.ResourceRequirements
	// Ensure we have a map[string]interface{} before passing it to the converter
	unstructuredMap, ok := in.(map[string]interface{})
	if !ok {
		return v1.ResourceRequirements{}, fmt.Errorf("unsupported resources type: %T", in)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredMap, &rr); err != nil {
		return v1.ResourceRequirements{}, fmt.Errorf("failed to convert to ResourceRequirements: %w", err)
	}
	return rr, nil
}

// UnstructuredToEnvVarList converts a generic interface (typically []interface{})
// representing a list of environment variables into a concrete []v1.EnvVar.
func UnstructuredToEnvVarList(in interface{}) ([]v1.EnvVar, error) {
	// If already the right type, return as-is
	if envVars, ok := in.([]v1.EnvVar); ok {
		return envVars, nil
	}

	slice, ok := in.([]interface{})
	if !ok || slice == nil {
		return []v1.EnvVar{}, fmt.Errorf("unsupported env var list type: %T", in)
	}

	var envVars []v1.EnvVar
	for _, item := range slice {
		m, ok := item.(map[string]interface{})
		if !ok || m == nil {
			slog.Warn("skipping unsupported env var item type", "type", item)
			continue
		}

		var ev v1.EnvVar
		if name, ok := m["name"].(string); ok {
			ev.Name = name
		}
		if value, ok := m["value"].(string); ok {
			ev.Value = value
		}

		// Handle ValueFrom
		if valueFromRaw, found := m["valueFrom"]; found && valueFromRaw != nil {
			ev.ValueFrom = &v1.EnvVarSource{}
			if valueFromMap, ok := valueFromRaw.(map[string]interface{}); ok {
				// Handle ConfigMapKeyRef
				if cmkRefRaw, found := valueFromMap["configMapKeyRef"]; found && cmkRefRaw != nil {
					if cmkRefMap, ok := cmkRefRaw.(map[string]interface{}); ok {
						ev.ValueFrom.ConfigMapKeyRef = &v1.ConfigMapKeySelector{}
						if name, ok := cmkRefMap["name"].(string); ok {
							ev.ValueFrom.ConfigMapKeyRef.Name = name
						}
						if key, ok := cmkRefMap["key"].(string); ok {
							ev.ValueFrom.ConfigMapKeyRef.Key = key
						}
						if optional, ok := cmkRefMap["optional"].(bool); ok {
							ev.ValueFrom.ConfigMapKeyRef.Optional = &optional
						}
					}
				}
				// Handle SecretKeyRef
				if skRefRaw, found := valueFromMap["secretKeyRef"]; found && skRefRaw != nil {
					if skRefMap, ok := skRefRaw.(map[string]interface{}); ok {
						ev.ValueFrom.SecretKeyRef = &v1.SecretKeySelector{}
						if name, ok := skRefMap["name"].(string); ok {
							ev.ValueFrom.SecretKeyRef.Name = name
						}
						if key, ok := skRefMap["key"].(string); ok {
							ev.ValueFrom.SecretKeyRef.Key = key
						}
						if optional, ok := skRefMap["optional"].(bool); ok {
							ev.ValueFrom.SecretKeyRef.Optional = &optional
						}
					}
				}
				// TODO: Add handling for FieldRef, ResourceFieldRef if needed
			}
		}

		envVars = append(envVars, ev)
	}

	return envVars, nil
}

// EnvVarListToUnstructured converts a slice of v1.EnvVar to a []interface{}
// suitable to be embedded in an unstructured object. Only name and value are
// preserved, consistent with UnstructuredToEnvVarList.
func EnvVarListToUnstructured(in []v1.EnvVar) []interface{} {
	if in == nil {
		return nil
	}
	out := make([]interface{}, 0, len(in))
	for _, ev := range in {
		m := map[string]interface{}{}
		if ev.Name != "" {
			m["name"] = ev.Name
		}
		if ev.Value != "" {
			m["value"] = ev.Value
		}

		if ev.ValueFrom != nil {
			valueFromMap := map[string]interface{}{}
			if ev.ValueFrom.ConfigMapKeyRef != nil {
				cmkRefMap := map[string]interface{}{}
				if ev.ValueFrom.ConfigMapKeyRef.Name != "" {
					cmkRefMap["name"] = ev.ValueFrom.ConfigMapKeyRef.Name
				}
				if ev.ValueFrom.ConfigMapKeyRef.Key != "" {
					cmkRefMap["key"] = ev.ValueFrom.ConfigMapKeyRef.Key
				}
				if ev.ValueFrom.ConfigMapKeyRef.Optional != nil {
					cmkRefMap["optional"] = *ev.ValueFrom.ConfigMapKeyRef.Optional
				}
				valueFromMap["configMapKeyRef"] = cmkRefMap
			}
			if ev.ValueFrom.SecretKeyRef != nil {
				skRefMap := map[string]interface{}{}
				if ev.ValueFrom.SecretKeyRef.Name != "" {
					skRefMap["name"] = ev.ValueFrom.SecretKeyRef.Name
				}
				if ev.ValueFrom.SecretKeyRef.Key != "" {
					skRefMap["key"] = ev.ValueFrom.SecretKeyRef.Key
				}
				if ev.ValueFrom.SecretKeyRef.Optional != nil {
					skRefMap["optional"] = *ev.ValueFrom.SecretKeyRef.Optional
				}
				valueFromMap["secretKeyRef"] = skRefMap
			}
			// TODO: Add handling for FieldRef, ResourceFieldRef if needed
			if len(valueFromMap) > 0 {
				m["valueFrom"] = valueFromMap
			}
		}
		out = append(out, m)
	}
	return out
}

// ResourceRequirementsToUnstructured converts v1.ResourceRequirements to a map[string]interface{}
// suitable to be embedded in an unstructured object.
func ResourceRequirementsToUnstructured(rr v1.ResourceRequirements) (map[string]interface{}, error) {
	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&rr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ResourceRequirements to unstructured: %w", err)
	}
	return unstructuredMap, nil
}

// ParseCassandraStrNodes reads spec.cassandra.nodes and returns a []interface{} suitable
// for unstructured.SetNestedSlice under spec.cassandra.connection.nodes.
// It accepts legacy comma-separated strings and returns a slice of ip/port maps.
func ParseCassandraStrNodes(oldSpec *unstructured.Unstructured) []interface{} {
	const defaultCassandraPort = 9042

	oldNodes, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "cassandra", "nodes")
	if err != nil {
		slog.Error("error retrieving cassandra nodes", "err", err)
		return []interface{}{}
	}
	if !found || oldNodes == nil {
		slog.Warn("spec.cassandra.nodes field is missing or empty in the input CR. Resulting CR will have no cassandra connection nodes.")
		return []interface{}{}
	}

	// Legacy: comma-separated string "host:port,host2:port"
	oldNodesStr, ok := oldNodes.(string)
	if !ok {
		slog.Warn("spec.cassandra.nodes has unsupported type; skipping nodes conversion")
		return []interface{}{}
	}

	// Parse the comma-separated string
	// Nodes are of the form host:port, multiple nodes separated by commas
	// e.g. "cassandra1.example.com:9042,cassandra2.example.com:9042"
	var nodes []interface{}
	for _, entry := range strings.Split(oldNodesStr, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		host, portStr, hasPort := strings.Cut(entry, ":")
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}

		// Default port if not specified
		port := int64(defaultCassandraPort)
		if hasPort {
			// Set port if valid
			if p, err := strconv.ParseInt(strings.TrimSpace(portStr), 10, 64); err == nil {
				port = p
			}
		}

		nodes = append(nodes, map[string]interface{}{
			"host": host,
			"port": port,
		})
	}

	return nodes
}

// MergeAdditionalEnv merges two additionalEnv fields (from API and Backend specs)
// Backend variables take precedence in case of conflicts (same name)
func MergeAdditionalEnv(apiEnv []v1.EnvVar, backendEnv []v1.EnvVar) []v1.EnvVar {
	// Delete duplicates from apiEnv
	for _, be := range backendEnv {
		for i, ae := range apiEnv {
			if ae.Name == be.Name {
				// Remove from apiEnv
				apiEnv = append(apiEnv[:i], apiEnv[i+1:]...)
				break
			}
		}
	}

	// Merge
	return append(apiEnv, backendEnv...)
}

// dumpResourceToYAMLFile dumps an unstructured resource to a YAML file at the given filepath.
func DumpResourceToYAMLFile(in *unstructured.Unstructured, filepath string) error {
	y, err := unstructuredToYAML(in)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, y, 0644)
}

// DumpYamlToUnstructured converts YAML to an unstructured object.
func DumpYamlToUnstructured(y []byte) (*unstructured.Unstructured, error) {
	var obj map[string]interface{}
	if err := yaml.Unmarshal(y, &obj); err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: obj}, nil
}

// unstructuredToJSON converts an unstructured object to JSON.
func unstructuredToJSON(in *unstructured.Unstructured) ([]byte, error) {
	out, err := json.Marshal(in.Object)
	if err != nil {
		return []byte{}, err
	}
	return out, nil
}

// unstructuredToYAML converts an unstructured object to YAML.
func unstructuredToYAML(in *unstructured.Unstructured) ([]byte, error) {
	j, err := unstructuredToJSON(in)
	if err != nil {
		return []byte{}, err
	}
	out, err := yaml.JSONToYAML(j)
	if err != nil {
		return []byte{}, err
	}
	return out, nil
}
