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
	"encoding/json"
	"fmt"
	"os"

	"github.com/astarte-platform/astartectl/config"
	"github.com/spf13/cobra"
)

var configImportCmd = &cobra.Command{
	Use:   "import <config_file>",
	Short: "Import configuration file",
	Long: `Import a configuration file previously exported with astartectl. This might
	be a partial export or a full export, and it should be in JSON format. By default, existing
	clusters and context won't be overwritten - you can force this behavior with --overwrite.`,
	Example: `  astartectl config import export.json`,
	Args:    cobra.ExactArgs(1),
	RunE:    configImportF,
}

var configExportCmd = &cobra.Command{
	Use:   "export [-o <output_file>]",
	Short: "Export current astartectl configuration",
	Long: `Export current astartectl configuration to a JSON file, which can be later imported through
	astartectl config import. By default, the complete list of clusters and contexts is exported, but you can
	tweak this behavior by using --clusters and --contexts, providing a list of names. By default, the resulting
	JSON file is printed on stdout, but it can also be saved to a file by specifying -o.`,
	Example: `  astartectl housekeeping realms create myrealm --realm-public-key /path/to/public_key`,
	Args:    cobra.ExactArgs(0),
	RunE:    configExportF,
}

func init() {
	ConfigCmd.AddCommand(configImportCmd)
	ConfigCmd.AddCommand(configExportCmd)

	configImportCmd.Flags().Bool("overwrite", false, "When specified, overwrites existing clusters or contexts with matching filenames")
	configImportCmd.Flags().StringSlice("clusters", []string{}, "A list of clusters to be imported, comma separated. If not specified, all clusters will be imported")
	configImportCmd.Flags().StringSlice("contexts", []string{}, "A list of contexts to be imported, comma separated. If not specified, all contexts will be imported")

	configExportCmd.Flags().StringP("output", "o", "", "If specified, configuration will be exported to specified file")
	configExportCmd.Flags().StringSlice("clusters", []string{}, "A list of clusters to be exported, comma separated. If not specified, all clusters will be exported")
	configExportCmd.Flags().StringSlice("contexts", []string{}, "A list of contexts to be exported, comma separated. If not specified, all contexts will be exported")
}

func configImportF(command *cobra.Command, args []string) error {
	overwrite, err := command.Flags().GetBool("overwrite")
	if err != nil {
		return err
	}
	clusters, err := command.Flags().GetStringSlice("clusters")
	if err != nil {
		return err
	}
	contexts, err := command.Flags().GetStringSlice("contexts")
	if err != nil {
		return err
	}

	// First of all, read the file content
	contents, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Load the bundle
	var bundle config.Bundle
	if err := json.Unmarshal(contents, &bundle); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Go for it
	if err := config.LoadBundleToDirectory(bundle, config.GetConfigDir(), clusters, contexts, overwrite); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Configuration imported successfully.")
	return nil
}

func configExportF(command *cobra.Command, args []string) error {
	output, err := command.Flags().GetString("output")
	if err != nil {
		return err
	}
	clusters, err := command.Flags().GetStringSlice("clusters")
	if err != nil {
		return err
	}
	contexts, err := command.Flags().GetStringSlice("contexts")
	if err != nil {
		return err
	}

	// Go for it
	bundle, err := config.CreateBundleFromDirectory(config.GetConfigDir(), clusters, contexts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Get some JSON out of it
	jsonBytes, err := json.Marshal(bundle)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if output == "" {
		fmt.Println(string(jsonBytes))
	} else {
		if err := os.WriteFile(output, jsonBytes, 0644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	return nil
}
