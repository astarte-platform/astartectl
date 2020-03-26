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
func CreateBundleFromDirectory(configDir string) (Bundle, error) {
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
		contextConfiguration, err := LoadContextConfiguration(configDir, context)
		if err != nil {
			return bundle, err
		}
		bundle.Contexts[context] = contextConfiguration
	}

	return bundle, nil
}

// LoadBundleToDirectory loads a bundle into a config directory, overwriting matching files, overwriting any
// overlapping entry. When failing, it might have partially loaded some of the entries in the file
func LoadBundleToDirectory(bundle Bundle, configDir string) error {
	// Load clusters first
	for name, configuration := range bundle.Clusters {
		if err := SaveClusterConfiguration(configDir, name, configuration); err != nil {
			return err
		}
	}

	// Then contexts
	for name, configuration := range bundle.Contexts {
		if err := SaveContextConfiguration(configDir, name, configuration); err != nil {
			return err
		}
	}

	// Then main conf
	return SaveBaseConfiguration(configDir, bundle.BaseConfig)
}
