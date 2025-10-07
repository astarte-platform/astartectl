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
// and the value is not nil.
//
// By preventing the addition of nil/zero values to the destination unstructured object when
// a field is missing, this helper sidesteps a severe issue in Kubernetes API where explicitly
// provided "null" entries in an unstructured map bypass schema defaulting hooks entirely.
func CopyIfExists(source *unstructured.Unstructured, dest *unstructured.Unstructured, sPath []string, dPath []string) error {
	if source == nil || dest == nil {
		return fmt.Errorf("source or dest is nil")
	}

	// Retrieve the field from source only if it exists
	sourceField, found, err := unstructured.NestedFieldNoCopy(source.Object, sPath...)
	if err != nil {
		return fmt.Errorf("error retrieving field %v from source: %w", sPath, err)
	}

	// If found and not nil, set it in dest
	if found && sourceField != nil {
		if err := unstructured.SetNestedField(dest.Object, sourceField, dPath...); err != nil {
			return fmt.Errorf("error setting field %v in dest: %w", dPath, err)
		}
	}

	return nil
}

// SumResourceRequirements sums two v1.ResourceRequirements objects.
// In the migration from v1alpha3 to v2alpha1, many components unified their independent
// `api` and `backend` pods into a monolithic deployment.
// To ensure the new monolith receives at least the same minimum scheduling bounds without
// degrading performance, limits and requests from both fragments must be strictly summated.
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
	var envVars []v1.EnvVar
	unstructuredSlice, ok := in.([]interface{})
	if !ok {
		// If it's already a typed slice, we can just convert it. This is a common case.
		if typedSlice, ok := in.([]v1.EnvVar); ok {
			return typedSlice, nil
		}
		return nil, fmt.Errorf("unsupported env var list type: %T", in)
	}

	for _, item := range unstructuredSlice {
		unstructuredMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid non-map item in env var list: %v", item)
		}
		var envVar v1.EnvVar
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredMap, &envVar); err != nil {
			return nil, fmt.Errorf("item cannot be converted to EnvVar: %w", err)
		}
		envVars = append(envVars, envVar)
	}

	return envVars, nil
}

// EnvVarListToUnstructured converts a slice of v1.EnvVar to a []interface{}
// suitable to be embedded in an unstructured object.
func EnvVarListToUnstructured(in []v1.EnvVar) ([]interface{}, error) {
	if in == nil {
		return nil, nil
	}
	out := make([]interface{}, len(in))
	for i, ev := range in {
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&ev)
		if err != nil {
			return nil, fmt.Errorf("failed to convert EnvVar to unstructured: %w", err)
		}
		out[i] = unstructuredMap
	}
	return out, nil
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
//
// In v1alpha3, Cassandra nodes were specified as a comma-separated string mapping host to port:
// "cassandra1.example.com:9042,cassandra2.example.com".
// In v2alpha1, this is represented as a structured list of maps:
// [{"host": "cassandra1.example.com", "port": 9042}, ...].
// This method handles the parsing and normalization of these strings into unstructured nodes list.
func ParseCassandraStrNodes(oldSpec *unstructured.Unstructured) ([]interface{}, error) {
	const defaultCassandraPort = 9042

	oldNodes, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "cassandra", "nodes")
	if err != nil {
		return nil, fmt.Errorf("error retrieving cassandra nodes: %w", err)
	}
	if !found || oldNodes == nil {
		slog.Warn("spec.cassandra.nodes field is missing or empty in the input CR. Resulting CR will have no cassandra connection nodes.")
		return []interface{}{}, nil
	}

	// Legacy: comma-separated string "host:port,host2:port"
	oldNodesStr, ok := oldNodes.(string)
	if !ok {
		return nil, fmt.Errorf("spec.cassandra.nodes has unsupported type %T; expected string", oldNodes)
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
			p, err := strconv.ParseInt(strings.TrimSpace(portStr), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid port %q for node %q: %w", portStr, host, err)
			}
			port = p
		}

		nodes = append(nodes, map[string]interface{}{
			"host": host,
			"port": port,
		})
	}

	return nodes, nil
}

// MergeAdditionalEnv merges two additionalEnv fields (from API and Backend specs)
// Backend variables take precedence in case of conflicts (same name)
func MergeAdditionalEnv(apiEnv []v1.EnvVar, backendEnv []v1.EnvVar) []v1.EnvVar {
	backendKeys := make(map[string]struct{}, len(backendEnv))
	for _, be := range backendEnv {
		backendKeys[be.Name] = struct{}{}
	}

	var merged []v1.EnvVar
	for _, ae := range apiEnv {
		if _, exists := backendKeys[ae.Name]; !exists {
			merged = append(merged, ae)
		}
	}

	merged = append(merged, backendEnv...)
	return merged
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
		return nil, fmt.Errorf("failed to marshal unstructured object to JSON: %w", err)
	}
	return out, nil
}

// unstructuredToYAML converts an unstructured object to YAML.
func unstructuredToYAML(in *unstructured.Unstructured) ([]byte, error) {
	j, err := unstructuredToJSON(in)
	if err != nil {
		return nil, err
	}
	out, err := yaml.JSONToYAML(j)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JSON to YAML: %w", err)
	}
	return out, nil
}
