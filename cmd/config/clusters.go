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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/astarte-platform/astartectl/config"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
)

var clustersCmd = &cobra.Command{
	Use:     "clusters",
	Short:   "Manage clusters",
	Long:    `List, show or create clusters in your astartectl configuration.`,
	Aliases: []string{"cluster"},
}

var clustersListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List clusters",
	Long:    "List clusters present in your astartectl configuration.",
	Example: `  astartectl config clusters list`,
	RunE:    clustersListF,
	Aliases: []string{"ls"},
}

var clustersShowCmd = &cobra.Command{
	Use:     "show <cluster_name>",
	Short:   "Show cluster",
	Long:    "Show a cluster in your astartectl configuration.",
	Example: `  astartectl config clusters show mycluster`,
	Args:    cobra.ExactArgs(1),
	RunE:    clustersShowF,
}

var clustersGetHousekeepingKeyCmd = &cobra.Command{
	Use:     "get-housekeeping-key <cluster_name>",
	Short:   "Get the Housekeeping key from a cluster",
	Long:    "Get the Housekeeping key from a cluster in your astartectl configuration. This will work only if a housekeeping key is set",
	Example: `  astartectl config clusters get-housekeeping-key mycluster`,
	Args:    cobra.ExactArgs(1),
	RunE:    clustersGetHousekeepingKeyF,
}

var clustersCreateCmd = &cobra.Command{
	Use:     "create <cluster_name>",
	Short:   "Create cluster",
	Long:    "Create a cluster in your astartectl configuration.",
	Example: `  astartectl config clusters create mycluster --api-url https://my.astarte.apis.example.com --housekeeping-key /path/to/private_key`,
	Args:    cobra.ExactArgs(1),
	RunE:    clustersCreateF,
}

var clustersUpdateCmd = &cobra.Command{
	Use:     "update <cluster_name>",
	Short:   "Update cluster",
	Long:    "Update a cluster in your astartectl configuration.",
	Example: `  astartectl config clusters update mycluster --api-url https://my.astarte.apis.example.com`,
	Args:    cobra.ExactArgs(1),
	RunE:    clustersUpdateF,
}

var clustersDeleteCmd = &cobra.Command{
	Use:     "delete <cluster_name>",
	Short:   "Delete cluster",
	Long:    "Delete a cluster in your Astarte instance.",
	Example: `  astartectl config clusters delete mycluster`,
	Args:    cobra.ExactArgs(1),
	RunE:    clustersDeleteF,
	Aliases: []string{"del"},
}

func init() {
	ConfigCmd.AddCommand(clustersCmd)

	clustersGetHousekeepingKeyCmd.Flags().StringP("output", "o", "", "If specified, housekeeping key will be saved to specified file")

	clustersCreateCmd.Flags().StringP("housekeeping-key", "k", "", "Path to PEM encoded private key used as housekeeping key")
	if err := clustersCreateCmd.MarkFlagFilename("housekeeping-key"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// TODO: Define the -t shorthand once we fix all the token everywhere mess
	clustersCreateCmd.Flags().String("housekeeping-token", "", "A JWT token used to authenticate against housekeeping. To be provided if key is not available")
	// TODO: Define the -u shorthand once we fix all the url everywhere mess
	clustersCreateCmd.Flags().String("api-url", "", "The base API URL for the Astarte Cluster")
	clustersCreateCmd.Flags().String("appengine-url", "", "The Appengine API URL for the Astarte Cluster")
	clustersCreateCmd.Flags().String("flow-url", "", "The Flow API URL for the Astarte Cluster")
	clustersCreateCmd.Flags().String("housekeeping-url", "", "The Housekeeping API URL for the Astarte Cluster")
	clustersCreateCmd.Flags().String("pairing-url", "", "The Pairing API URL for the Astarte Cluster")
	clustersCreateCmd.Flags().String("realm-management-url", "", "The Realm Management API URL for the Astarte Cluster")

	clustersUpdateCmd.Flags().StringP("housekeeping-key", "k", "", "Path to PEM encoded private key used as housekeeping key")
	if err := clustersUpdateCmd.MarkFlagFilename("housekeeping-key"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// TODO: Define the -t shorthand once we fix all the token everywhere mess
	clustersUpdateCmd.Flags().String("housekeeping-token", "", "A JWT token used to authenticate against housekeeping. To be provided if key is not available")
	// TODO: Define the -u shorthand once we fix all the url everywhere mess
	clustersUpdateCmd.Flags().String("api-url", "", "The base API URL for the Astarte Cluster")
	clustersUpdateCmd.Flags().String("appengine-url", "", "The Appengine API URL for the Astarte Cluster")
	clustersUpdateCmd.Flags().String("flow-url", "", "The Flow API URL for the Astarte Cluster")
	clustersUpdateCmd.Flags().String("housekeeping-url", "", "The Housekeeping API URL for the Astarte Cluster")
	clustersUpdateCmd.Flags().String("pairing-url", "", "The Pairing API URL for the Astarte Cluster")
	clustersUpdateCmd.Flags().String("realm-management-url", "", "The Realm Management API URL for the Astarte Cluster")

	clustersCmd.AddCommand(
		clustersListCmd,
		clustersShowCmd,
		clustersGetHousekeepingKeyCmd,
		clustersCreateCmd,
		clustersUpdateCmd,
		clustersDeleteCmd,
	)
}

func clustersListF(command *cobra.Command, args []string) error {
	configDir := config.GetConfigDir()
	clusters, err := config.ListClusterConfigurations(configDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', tabwriter.DiscardEmptyColumns)
	// header
	fmt.Fprintln(w, "CLUSTER NAME\tAPI URL(S)")
	for _, c := range clusters {
		cluster, err := config.LoadClusterConfiguration(configDir, c)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Fprintf(w, "%s\t", c)
		if cluster.URL != "" {
			fmt.Fprintln(w, cluster.URL)
		} else {
			fmt.Fprintf(w, "Appengine: %s\n", cluster.IndividualURLs.AppEngine)
			fmt.Fprintf(w, "\tFlow: %s\n", cluster.IndividualURLs.Flow)
			fmt.Fprintf(w, "\tHousekeeping: %s\n", cluster.IndividualURLs.Housekeeping)
			fmt.Fprintf(w, "\tPairing: %s\n", cluster.IndividualURLs.Pairing)
			fmt.Fprintf(w, "\tRealm Management: %s\n", cluster.IndividualURLs.RealmManagement)
		}
	}

	w.Flush()
	return nil
}

func clustersShowF(command *cobra.Command, args []string) error {
	clusterName := args[0]

	cluster, err := config.LoadClusterConfiguration(config.GetConfigDir(), clusterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	if cluster.URL != "" {
		fmt.Fprintf(w, "Astarte API URL:\t%s\n", cluster.URL)
	} else {
		fmt.Fprintf(w, "Astarte Individual API URLs:\t")
		fmt.Fprintf(w, "Appengine: %s\n", cluster.IndividualURLs.AppEngine)
		fmt.Fprintf(w, "\tFlow: %s\n", cluster.IndividualURLs.Flow)
		fmt.Fprintf(w, "\tHousekeeping: %s\n", cluster.IndividualURLs.Housekeeping)
		fmt.Fprintf(w, "\tPairing: %s\n", cluster.IndividualURLs.Pairing)
		fmt.Fprintf(w, "\tRealm Management: %s\n", cluster.IndividualURLs.RealmManagement)
	}
	switch {
	case cluster.Housekeeping.Key != "":
		fmt.Fprintln(w, "Housekeeping Authentication:\tKey")
	case cluster.Housekeeping.Token != "":
		fmt.Fprintln(w, "Housekeeping Authentication:\tToken")
	default:
		fmt.Fprintln(w, "Housekeeping Authentication:\tNone")
	}
	w.Flush()

	return nil
}

func clustersGetHousekeepingKeyF(command *cobra.Command, args []string) error {
	clusterName := args[0]
	output, err := command.Flags().GetString("output")
	if err != nil {
		return err
	}

	cluster, err := config.LoadClusterConfiguration(config.GetConfigDir(), clusterName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if cluster.Housekeeping.Key == "" {
		fmt.Fprintf(os.Stderr, "Cluster %s has no Housekeeping Key associated\n", clusterName)
	}

	decoded, err := base64.StdEncoding.DecodeString(cluster.Housekeeping.Key)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if output != "" {
		// Save to file
		if err := ioutil.WriteFile(output, decoded, 0644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		// Just throw it to stdout
		fmt.Println(string(decoded))
	}

	return nil
}

func clustersCreateF(command *cobra.Command, args []string) error {
	return performClusterCreation(args[0], false, command, args)
}

func clustersUpdateF(command *cobra.Command, args []string) error {
	return performClusterCreation(args[0], true, command, args)
}

func clustersDeleteF(command *cobra.Command, args []string) error {
	clusterName := args[0]

	if _, err := config.LoadClusterConfiguration(config.GetConfigDir(), clusterName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if ok, err := utils.AskForConfirmation(fmt.Sprintf("Will delete cluster %s. Are you sure you want to continue?", clusterName)); !ok || err != nil {
		return nil
	}

	if err := config.DeleteClusterConfiguration(config.GetConfigDir(), clusterName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Cluster %s deleted successfully\n", clusterName)
	return nil
}

func performClusterCreation(clusterName string, isUpdate bool, command *cobra.Command, args []string) error {
	configDir := config.GetConfigDir()
	housekeepingKey, err := command.Flags().GetString("housekeeping-key")
	if err != nil {
		return err
	}
	housekeepingToken, err := command.Flags().GetString("housekeeping-token")
	if err != nil {
		return err
	}
	apiURL, err := command.Flags().GetString("api-url")
	if err != nil {
		return err
	}
	appengineURL, err := command.Flags().GetString("appengine-url")
	if err != nil {
		return err
	}
	flowURL, err := command.Flags().GetString("flow-url")
	if err != nil {
		return err
	}
	housekeepingURL, err := command.Flags().GetString("housekeeping-url")
	if err != nil {
		return err
	}
	pairingURL, err := command.Flags().GetString("pairing-url")
	if err != nil {
		return err
	}
	realmManagementURL, err := command.Flags().GetString("realm-management-url")
	if err != nil {
		return err
	}

	individualURLsSpecified := appengineURL != "" || flowURL != "" || housekeepingURL != "" || pairingURL != "" || realmManagementURL != ""

	// Sanity checks
	switch {
	case apiURL != "" && individualURLsSpecified:
		fmt.Fprintln(os.Stderr, "You cannot specify both --api-url and individual URLs")
		os.Exit(1)
	case apiURL == "" && !individualURLsSpecified && !isUpdate:
		fmt.Fprintln(os.Stderr, "You should either specify --api-url or any number of individual URLs")
		os.Exit(1)
	case housekeepingKey != "" && housekeepingToken != "":
		fmt.Fprintln(os.Stderr, "You should specify only one among --housekeeping-key and --housekeeping-token")
		os.Exit(1)
	}

	// Check on the cluster
	cluster, err := config.LoadClusterConfiguration(configDir, clusterName)
	switch {
	case err != nil && isUpdate:
		fmt.Fprintf(os.Stderr, "Cluster %s does not exist. Maybe you meant to use create?\n", clusterName)
		os.Exit(1)
	case err == nil && !isUpdate:
		fmt.Fprintf(os.Stderr, "A cluster named %s already exists. Maybe you meant to use update?\n", clusterName)
		os.Exit(1)
	}

	// Go
	if apiURL != "" {
		cluster.URL = apiURL
	}
	if individualURLsSpecified {
		if appengineURL != "" {
			cluster.IndividualURLs.AppEngine = appengineURL
		}
		if flowURL != "" {
			cluster.IndividualURLs.Flow = flowURL
		}
		if housekeepingURL != "" {
			cluster.IndividualURLs.Housekeeping = housekeepingURL
		}
		if pairingURL != "" {
			cluster.IndividualURLs.Pairing = pairingURL
		}
		if realmManagementURL != "" {
			cluster.IndividualURLs.RealmManagement = realmManagementURL
		}
	}
	if housekeepingKey != "" {
		contents, err := ioutil.ReadFile(housekeepingKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		cluster.Housekeeping.Key = base64.StdEncoding.EncodeToString(contents)
		cluster.Housekeeping.Token = ""
	}
	if housekeepingToken != "" {
		cluster.Housekeeping.Token = housekeepingToken
		cluster.Housekeeping.Key = ""
	}

	// Save
	if err := config.SaveClusterConfiguration(configDir, clusterName, cluster, true); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Cluster %s saved successfully\n", clusterName)
	return nil
}
