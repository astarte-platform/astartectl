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
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

// IndividualURLsConfiguration represents configuration for Individual URLs.
// When present, URL is overridden and not used.
type IndividualURLsConfiguration struct {
	// AppEngine is AppEngine API URL
	AppEngine string `yaml:"appengine,omitempty" json:"appengine,omitempty"`
	// Flow is Flow API URL
	Flow string `yaml:"flow,omitempty" json:"flow,omitempty"`
	// Housekeeping is Housekeeping API URL
	Housekeeping string `yaml:"housekeeping,omitempty" json:"housekeeping,omitempty"`
	// Pairing is Pairing API URL
	Pairing string `yaml:"pairing,omitempty" json:"pairing,omitempty"`
	// RealmManagement is Realm Management API URL
	RealmManagement string `yaml:"realm-management,omitempty" json:"realm-management,omitempty"`
}

// HousekeepingConfiguration represents configuration for housekeeping
type HousekeepingConfiguration struct {
	// Key is the private key of the realm, encoded in base64
	Key string `yaml:"key,omitempty" json:"key,omitempty"`
	// Token is a token used to authenticate against the realm. When set, it takes precedence over key
	Token string `yaml:"token,omitempty" json:"token,omitempty"`
}

// ClusterFile represents a Cluster file
type ClusterFile struct {
	// URL is the base API URL for the Cluster. Can be omitted when specifying individual URLs
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
	// Housekeeping is the Cluster's Housekeeping configuration. Can be omitted if there is no such information
	Housekeeping HousekeepingConfiguration `yaml:"housekeeping,omitempty" json:"housekeeping,omitempty"`
	// IndividualURLs are the Cluster's individual API URLs. Should usually be omitted, and URL should
	// be used instead, but in case of exotic API server setups they are required. When specified, they
	// take precedence over URL
	IndividualURLs IndividualURLsConfiguration `yaml:"individual-urls,omitempty" json:"individual-urls,omitempty"`
}

// ListClusterConfigurations returns a list of available cluster configurations
func ListClusterConfigurations(configDir string) ([]string, error) {
	return listYamlNames(clustersDirFromConfigDir(configDir))
}

// LoadClusterConfiguration loads a cluster configuration from the config directory
func LoadClusterConfiguration(configDir, clusterName string) (ClusterFile, error) {
	cluster := ClusterFile{}
	contents, err := loadYamlFile(clustersDirFromConfigDir(configDir), clusterName)
	if err != nil {
		return cluster, err
	}
	err = yaml.Unmarshal(contents, &cluster)
	return cluster, err
}

// SaveClusterConfiguration saves a cluster configuration in the config directory
func SaveClusterConfiguration(configDir, clusterName string, configuration ClusterFile, overwrite bool) error {
	configPath := path.Join(clustersDirFromConfigDir(configDir), clusterName+".yaml")

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

// DeleteClusterConfiguration deletes a cluster configuration in the config directory. It will return
// an error if the cluster does not exist. The operation cannot be reverted
func DeleteClusterConfiguration(configDir, clusterName string) error {
	fileName, err := getYamlFilename(clustersDirFromConfigDir(configDir), clusterName)
	if err != nil {
		return err
	}

	return os.Remove(fileName)
}
