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

import (
	"gopkg.in/yaml.v2"
	"os"
	"path"
)

const baseConfigName = "astartectl"

// BaseConfigFile represents a base configuration for astartectl
type BaseConfigFile struct {
	// CurrentContext represents the context which should be used when no context is explicitly specified
	CurrentContext string `yaml:"context" json:"context"`
}

// LoadBaseConfiguration loads the base configuration from a config directory
func LoadBaseConfiguration(configDir string) (BaseConfigFile, error) {
	if configDir == "" {
		configDir = GetDefaultConfigDir()
	}

	config := BaseConfigFile{}
	contents, err := loadYamlFile(configDir, baseConfigName)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(contents, &config)
	return config, err
}

// SaveBaseConfiguration saves the base configuration in a config directory
func SaveBaseConfiguration(configDir string, configuration BaseConfigFile) error {
	if configDir == "" {
		configDir = GetDefaultConfigDir()
	}

	if err := ensureConfigDirectoryStructure(configDir); err != nil {
		return err
	}

	contents, err := yaml.Marshal(configuration)
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(configDir, baseConfigName+".yaml"), contents, 0644)
}
