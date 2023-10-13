// Copyright © 2019 Ispirata Srl
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
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

// RealmConfiguration represents configuration for a Realm
type RealmConfiguration struct {
	// Name is the name of the realm
	Name string `yaml:"name" json:"name"`
	// Key is the private key of the realm, encoded in base64. Can be omitted
	Key string `yaml:"key,omitempty" json:"key,omitempty"`
	// Token is a token used to authenticate against the realm. When set, it takes precedence over key
	Token string `yaml:"token,omitempty" json:"token,omitempty"`
}

// ContextFile represents a Context file
type ContextFile struct {
	// Cluster is the name of the cluster. It has to match an existing Cluster configuration
	Cluster string `yaml:"cluster" json:"cluster"`
	// Realm is the realm object. In case the Context refers to Housekeeping only, can be omitted
	Realm RealmConfiguration `yaml:"realm,omitempty" json:"realm,omitempty"`
}

// ListContextConfigurations returns a list of available context configurations
func ListContextConfigurations(configDir string) ([]string, error) {
	return listYamlNames(contextsDirFromConfigDir(configDir))
}

// LoadContextConfiguration loads a context configuration from the config directory
func LoadContextConfiguration(configDir, contextName string) (ContextFile, error) {
	context := ContextFile{}
	contents, err := loadYamlFile(contextsDirFromConfigDir(configDir), contextName)
	if err != nil {
		return context, err
	}
	err = yaml.Unmarshal(contents, &context)
	return context, err
}

// SaveContextConfiguration saves a context configuration in the config directory
func SaveContextConfiguration(configDir, contextName string, configuration ContextFile, overwrite bool) error {
	configPath := path.Join(contextsDirFromConfigDir(configDir), contextName+".yaml")

	if !overwrite {
		if _, err := os.Stat(configPath); err == nil {
			// Don't overwrite, don't fail
			return nil
		}
	}

	if err := ensureConfigDirectoryStructure(configDir); err != nil {
		return err
	}

	contents, err := yaml.Marshal(configuration)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, contents, 0644)
}

// DeleteContextConfiguration deletes a context configuration in the config directory. It will return
// an error if the context does not exist. The operation cannot be reverted
func DeleteContextConfiguration(configDir, contextName string) error {
	fileName, err := getYamlFilename(contextsDirFromConfigDir(configDir), contextName)
	if err != nil {
		return err
	}

	return os.Remove(fileName)
}
