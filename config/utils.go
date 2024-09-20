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
	"fmt"
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
	currentContextInterface := viper.Get("context")
	currentContext := ""
	//in case of empty file, the following check used to fail, now it defaults to empty context and enable file rewrite
	if currentContextInterface != nil {
		currentContext = currentContextInterface.(string)
	}

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
	contextViper := viper.New()
	contextViper.SetConfigName(currentContext)
	contextViper.AddConfigPath(contextsDirFromConfigDir(configDir))
	if err := contextViper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.MergeConfigMap(contextViper.AllSettings()); err != nil {
		return err
	}

	// Now, get the corresponding cluster
	cluster := viper.Get("cluster").(string)
	if cluster == "" {
		return errors.New("No cluster defined in context - something is wrong")
	}

	// Load the corresponding cluster
	clusterViper := viper.New()
	clusterViper.SetConfigName(cluster)
	clusterViper.AddConfigPath(clustersDirFromConfigDir(configDir))
	if err := clusterViper.ReadInConfig(); err != nil {
		return err
	}

	// Done loading
	return viper.MergeConfigMap(clusterViper.AllSettings())
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

	if os.IsNotExist(err) {
		//if we cannot open the directory, create it and open again
		_ = os.MkdirAll(dirName, 0755)
		file, err = os.Open(dirName)
		if err != nil {
			return nil, err
		}

	} else if err != nil {
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

func getYamlFilename(dirName, fileName string) (string, error) {
	if _, err := os.Stat(path.Join(dirName, fileName+".yaml")); err == nil {
		return path.Join(dirName, fileName+".yaml"), nil
	} else if _, err := os.Stat(path.Join(dirName, fileName+".yml")); err == nil {
		return path.Join(dirName, fileName+".yml"), nil
	}
	return "", os.ErrNotExist
}

func loadYamlFile(dirName, fileName string) ([]byte, error) {
	yamlFileName, err := getYamlFilename(dirName, fileName)
	if err != nil {
		return nil, err
	}

	// Read file contents and return it
	return os.ReadFile(yamlFileName)
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

func GetBaseConfig(configDir string) BaseConfigFile {

	baseConfig, err := LoadBaseConfiguration(configDir)
	if err != nil {
		// Shoot out a warning, but don't fail
		baseConfig = BaseConfigFile{}
		fmt.Fprintf(os.Stderr, "warn: Could not load configuration file: %s. Will proceed creating a new one\n", err.Error())
		// Now set the current context to the new one
		baseConfig.CurrentContext = ""

		if err := SaveBaseConfiguration(configDir, baseConfig); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	return baseConfig
}
func UpdateBaseConfigWithContext(configDir, context string) BaseConfigFile {
	baseConfig, err := LoadBaseConfiguration(configDir)
	if err != nil {
		// Shoot out a warning, but don't fail
		baseConfig = BaseConfigFile{}
		fmt.Fprintf(os.Stderr, "warn: Could not load configuration file: %s. Will proceed creating a new one\n", err.Error())
		// Now set the current context to the new one
	}

	baseConfig.CurrentContext = context

	if err := SaveBaseConfiguration(configDir, baseConfig); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return baseConfig
}
