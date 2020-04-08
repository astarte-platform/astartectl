// Copyright © 2020 Ispirata Srl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

// Bundle represents a bundle encapsulating all configuration files
type Bundle struct {
	// BaseConfig is the base config for astartectl
	BaseConfig BaseConfigFile `yaml:"config" json:"config"`
	// Clusters is the list of available clusters
	Clusters map[string]ClusterFile `yaml:"clusters" json:"clusters"`
	// Contexts is the list of available contexts
	Contexts map[string]ContextFile `yaml:"contexts" json:"contexts"`
}

// CreateBundleFromDirectory returns a bundle out of a config directory
// Clusters and Contexts, when specified, are a list of clusters and contexts to save.
// An empty list means "all"
func CreateBundleFromDirectory(configDir string, clusters, contexts []string) (Bundle, error) {
	bundle := Bundle{
		BaseConfig: BaseConfigFile{},
		Clusters:   map[string]ClusterFile{},
		Contexts:   map[string]ContextFile{},
	}

	var err error
	// Load main config
	if bundle.BaseConfig, err = LoadBaseConfiguration(configDir); err != nil {
		return bundle, nil
	}

	// Load Clusters
	clustersList, err := ListClusterConfigurations(configDir)
	if err != nil {
		return bundle, nil
	}

	for _, cluster := range clustersList {
		if len(clusters) > 0 {
			if !existsInStringSlice(cluster, clusters) {
				// skip
				continue
			}
		}
		clusterConfiguration, err := LoadClusterConfiguration(configDir, cluster)
		if err != nil {
			return bundle, err
		}
		bundle.Clusters[cluster] = clusterConfiguration
	}

	// Load Contexts
	contextsList, err := ListContextConfigurations(configDir)
	if err != nil {
		return bundle, nil
	}

	for _, context := range contextsList {
		if len(contexts) > 0 {
			if !existsInStringSlice(context, contexts) {
				// skip
				continue
			}
		}
		contextConfiguration, err := LoadContextConfiguration(configDir, context)
		if err != nil {
			return bundle, err
		}
		bundle.Contexts[context] = contextConfiguration
	}

	return bundle, nil
}

// LoadBundleToDirectory loads a bundle into a config directory, overwriting matching files, overwriting any
// overlapping entry. When failing, it might have partially loaded some of the entries in the file.
// Clusters and Contexts, when specified, are a list of clusters and contexts to load.
// An empty list means "all"
func LoadBundleToDirectory(bundle Bundle, configDir string, clusters, contexts []string, overwrite bool) error {
	// Load clusters first
	for name, configuration := range bundle.Clusters {
		if len(clusters) > 0 {
			if !existsInStringSlice(name, clusters) {
				// skip
				continue
			}
		}
		if err := SaveClusterConfiguration(configDir, name, configuration, overwrite); err != nil {
			return err
		}
	}

	// Then contexts
	for name, configuration := range bundle.Contexts {
		if len(contexts) > 0 {
			if !existsInStringSlice(name, contexts) {
				// skip
				continue
			}
		}
		if err := SaveContextConfiguration(configDir, name, configuration, overwrite); err != nil {
			return err
		}
	}

	// Then main conf
	return SaveBaseConfiguration(configDir, bundle.BaseConfig)
}

func existsInStringSlice(match string, list []string) bool {
	for _, v := range list {
		if v == match {
			return true
		}
	}
	return false
}
