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
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/shibukawa/configdir"
	"github.com/spf13/viper"
)

// ConfigureViper sets up Viper to behave correctly with regards to both context and
// configuration directory, taking into account all environment variables and parameters.
// Order of precedence is: override, environment variables, defaults
func ConfigureViper(contextOverride string) error {
	// Get configuration directory, first of all
	configDir := GetConfigDir()
	// Check if it exists
	if _, err := os.Stat(configDir); err != nil {
		return err
	}

	// Load base config first of all
	viper.SetConfigName(baseConfigName)
	viper.AddConfigPath(configDir)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	// Now, get the current context
	currentContext := viper.Get("context").(string)
	// Check overrides
	if contextOverride != "" {
		currentContext = contextOverride
	} else if contextFromEnv, ok := os.LookupEnv("ASTARTE_CONTEXT"); ok && contextFromEnv != "" {
		currentContext = contextFromEnv
	}

	if currentContext == "" {
		return errors.New("No current context defined")
	}

	// Load the current context
	viper.SetConfigName(currentContext)
	viper.AddConfigPath(contextsDirFromConfigDir(configDir))
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	// Now, get the corresponding cluster
	cluster := viper.Get("cluster").(string)
	if cluster == "" {
		return errors.New("No cluster defined in context - something is wrong")
	}

	// Load the corresponding cluster
	viper.SetConfigName(cluster)
	viper.AddConfigPath(clustersDirFromConfigDir(configDir))
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	// Done loading
	return nil
}

// GetConfigDir returns the Config Dir based on the current status
func GetConfigDir() string {
	configDir := GetDefaultConfigDir()
	if configDirOverride := viper.GetString("config-dir"); configDirOverride != "" {
		configDir = configDirOverride
	} else if dirFromEnv, ok := os.LookupEnv("ASTARTE_CONFIG_DIR"); ok && dirFromEnv != "" {
		configDir = dirFromEnv
	}
	return configDir
}

// GetDefaultConfigDir returns the default config directory
func GetDefaultConfigDir() string {
	// TODO: In the future, we might want to have proper system vs. local configuration, but
	// for now we're just assuming we care about local configuration, and that's it.
	return configdir.New("", "astarte").QueryFolders(configdir.Global)[0].Path
}

func clustersDirFromConfigDir(configDir string) string {
	if configDir == "" {
		configDir = GetConfigDir()
	}
	return path.Join(configDir, "clusters")
}

func contextsDirFromConfigDir(configDir string) string {
	if configDir == "" {
		configDir = GetConfigDir()
	}
	return path.Join(configDir, "contexts")
}

func listYamlNames(dirName string) ([]string, error) {
	file, err := os.Open(dirName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// This is a lot faster than ioutil.ReadDir, and we don't need much beyond the name.
	list, err := file.Readdirnames(0) // 0 to read all files and folders
	if err != nil {
		return nil, err
	}

	names := []string{}
	for _, name := range list {
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			// skip
			continue
		}
		// Get the file name
		fileName := strings.TrimSuffix(name, ".yaml")
		fileName = strings.TrimSuffix(fileName, ".yml")
		// Append it
		names = append(names, fileName)
	}

	return names, nil
}

func loadYamlFile(dirName, fileName string) ([]byte, error) {
	yamlFileName := ""
	if _, err := os.Stat(path.Join(dirName, fileName+".yaml")); err != nil {
		yamlFileName = path.Join(dirName, fileName+".yaml")
	} else if _, err := os.Stat(path.Join(dirName, fileName+".yml")); err != nil {
		yamlFileName = path.Join(dirName, fileName+".yml")
	} else {
		return nil, os.ErrNotExist
	}

	// Read file contents and return it
	return ioutil.ReadFile(yamlFileName)
}

func ensureConfigDirectoryStructure(configDir string) error {
	if configDir == "" {
		configDir = GetDefaultConfigDir()
	}
	dirsToEnsure := []string{configDir, clustersDirFromConfigDir(configDir), contextsDirFromConfigDir(configDir)}
	for _, d := range dirsToEnsure {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			if err := os.Mkdir(d, os.ModePerm); err != nil {
				return err
			}
		}
	}
	return nil
}
