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

package cluster

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"

	"github.com/astarte-platform/astartectl/config"
	"github.com/spf13/cobra"
)

var getClusterConfigCmd = &cobra.Command{
	Use:   "get-cluster-config <name>",
	Short: "Gets the current cluster config for instance <name> and updates your local Astarte configuration",
	Long: `Fetches the current cluster config for instance <name>, including the Cluster URL and its
housekeeping key, and adds it to your cluster and context configuration. A new cluster entry
will be created, together with a new context matching the cluster, without an associated realm.`,
	Example: `  astartectl cluster instances get-cluster-config`,
	RunE:    instancesGetClusterConfigF,
	Args:    cobra.ExactArgs(1),
}

func init() {
	getClusterConfigCmd.PersistentFlags().String("namespace", "astarte", "Namespace in which to look for the Astarte resource.")

	InstancesCmd.AddCommand(getClusterConfigCmd)
}

func instancesGetClusterConfigF(command *cobra.Command, args []string) error {
	resourceName := args[0]
	resourceNamespace, err := command.Flags().GetString("namespace")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if resourceNamespace == "" {
		resourceNamespace = "astarte"
	}

	return doGetClusterConfig(resourceName, resourceNamespace)
}

func doGetClusterConfig(resourceName, resourceNamespace string) error {
	astarteObject, err := getAstarteInstance(resourceName, resourceNamespace)
	if err != nil {
		fmt.Printf("Error while looking for instance %s: %s.\n", resourceName, err.Error())
		os.Exit(1)
	}
	astarteSpec := astarteObject.Object["spec"].(map[string]interface{})

	astarteHost := astarteSpec["api"].(map[string]interface{})["host"].(string)
	astarteURL := url.URL{
		Host:   astarteSpec["api"].(map[string]interface{})["host"].(string),
		Scheme: "https",
	}

	clusterName := fmt.Sprintf("%s-%s-cluster", resourceName, astarteHost)

	// Fetch key
	keyData, err := getHousekeepingKey(resourceName, resourceNamespace, false)
	if err != nil {
		fmt.Printf("Error while fetching Housekeeping Key for instance %s: %s.\n", resourceName, err.Error())
		os.Exit(1)
	}

	clusterConfig := config.ClusterFile{
		URL: astarteURL.String(),
		Housekeeping: config.HousekeepingConfiguration{
			Key: base64.StdEncoding.EncodeToString(keyData),
		},
	}

	configDir := config.GetConfigDir()

	if err := config.SaveClusterConfiguration(configDir, clusterName, clusterConfig, true); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Created new Cluster configuration %s\n", clusterName)

	// Add a Context now
	contextName := fmt.Sprintf("%s-%s-global", resourceName, astarteHost)
	contextConfig := config.ContextFile{
		Cluster: clusterName,
	}
	if err := config.SaveContextConfiguration(configDir, contextName, contextConfig, true); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Created new Context %s\n", contextName)

	// Now set the current context to the new one
	baseConfig, err := config.LoadBaseConfiguration(configDir)
	if err != nil {
		// Shoot out a warning, but don't fail
		baseConfig = config.BaseConfigFile{}
		fmt.Fprintf(os.Stderr, "warn: Could not load configuration file: %s. Will proceed creating a new one\n", err.Error())
	}

	baseConfig.CurrentContext = contextName
	if err := config.SaveBaseConfiguration(configDir, baseConfig); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Context switched to %s\n", contextName)

	return nil
}
