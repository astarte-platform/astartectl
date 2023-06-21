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
	"os"
	"text/tabwriter"

	"github.com/astarte-platform/astartectl/config"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
)

var contextsCmd = &cobra.Command{
	Use:     "contexts",
	Short:   "Manage contexts",
	Long:    `List, show or create contexts in your astartectl configuration.`,
	Aliases: []string{"context"},
}

var contextsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List contexts",
	Long:    "List contexts present in your astartectl configuration.",
	Example: `  astartectl config contexts list`,
	RunE:    contextsListF,
	Aliases: []string{"ls"},
}

var contextsShowCmd = &cobra.Command{
	Use:     "show <context_name>",
	Short:   "Show context",
	Long:    "Show a context in your astartectl configuration.",
	Example: `  astartectl config contexts show mycontext`,
	Args:    cobra.ExactArgs(1),
	RunE:    contextsShowF,
}

var contextsGetRealmKeyCmd = &cobra.Command{
	Use:     "get-realm-key <context_name>",
	Short:   "Get the Realm key from a context",
	Long:    "Get the Realm key from a context in your astartectl configuration. This will work only if a realm key is set",
	Example: `  astartectl config contexts get-realm-key mycontext`,
	Args:    cobra.ExactArgs(1),
	RunE:    contextsGetRealmKeyF,
}

var contextsCreateCmd = &cobra.Command{
	Use:     "create <context_name>",
	Short:   "Create context",
	Long:    "Create a context in your astartectl configuration.",
	Example: `  astartectl config contexts create mycontext --cluster mycluster --realm-name myrealm --realm-key /path/to/private_key`,
	Args:    cobra.ExactArgs(1),
	RunE:    contextsCreateF,
}

var contextsUpdateCmd = &cobra.Command{
	Use:     "update <context_name>",
	Short:   "Update context",
	Long:    "Update a context in your astartectl configuration.",
	Example: `  astartectl config contexts update mycontext --realm-key /path/to/private_key`,
	Args:    cobra.ExactArgs(1),
	RunE:    contextsUpdateF,
}

var contextsDeleteCmd = &cobra.Command{
	Use:     "delete <context_name>",
	Short:   "Delete context",
	Long:    "Delete a context in your Astarte instance.",
	Example: `  astartectl config contexts delete mycontext`,
	Args:    cobra.ExactArgs(1),
	RunE:    contextsDeleteF,
	Aliases: []string{"del"},
}

func init() {
	ConfigCmd.AddCommand(contextsCmd)

	contextsGetRealmKeyCmd.Flags().StringP("output", "o", "", "If specified, private key will be saved to specified file")

	contextsCreateCmd.Flags().StringP("realm-private-key", "k", "", "Path to PEM encoded private key used as realm key")
	if err := contextsCreateCmd.MarkFlagFilename("realm-private-key"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	contextsCreateCmd.Flags().StringP("realm-name", "r", "", "Realm name to which the context should be associated to")
	// TODO: Define the -t shorthand once we fix all the token everywhere mess
	contextsCreateCmd.Flags().String("realm-token", "", "A JWT token used to authenticate against the realm. To be provided if key is not available")
	contextsCreateCmd.Flags().StringP("cluster", "c", "", "The cluster name the context should refer to. Must be an existing astartectl cluster")
	if err := contextsCreateCmd.MarkFlagRequired("cluster"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	contextsCreateCmd.Flags().BoolP("activate", "a", false, "When specified, activates the context upon its creation")

	contextsUpdateCmd.Flags().StringP("realm-private-key", "k", "", "Path to PEM encoded private key used as realm key")
	if err := contextsUpdateCmd.MarkFlagFilename("realm-private-key"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	contextsUpdateCmd.Flags().StringP("realm-name", "r", "", "Realm name to which the context should be associated to")
	// TODO: Define the -t shorthand once we fix all the token everywhere mess
	contextsUpdateCmd.Flags().String("realm-token", "", "A JWT token used to authenticate against the realm. To be provided if key is not available")
	contextsUpdateCmd.Flags().StringP("cluster", "c", "", "The cluster name the context should refer to. Must be an existing astartectl cluster")
	contextsUpdateCmd.Flags().BoolP("activate", "a", false, "When specified, activates the context after updating it")

	contextsCmd.AddCommand(
		contextsListCmd,
		contextsShowCmd,
		contextsGetRealmKeyCmd,
		contextsCreateCmd,
		contextsUpdateCmd,
		contextsDeleteCmd,
	)
}

func contextsListF(command *cobra.Command, args []string) error {
	configDir := config.GetConfigDir()
	contexts, err := config.ListContextConfigurations(configDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	baseConfig, err := config.LoadBaseConfiguration(config.GetConfigDir())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', tabwriter.DiscardEmptyColumns)
	// header
	fmt.Fprintln(w, "CONTEXT NAME\tCLUSTER\tREALM NAME")
	for _, c := range contexts {
		context, err := config.LoadContextConfiguration(configDir, c)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if baseConfig.CurrentContext == c {
			fmt.Fprint(w, "* ")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", c, context.Cluster, context.Realm.Name)
	}

	w.Flush()
	return nil
}

func contextsShowF(command *cobra.Command, args []string) error {
	contextName := args[0]

	context, err := config.LoadContextConfiguration(config.GetConfigDir(), contextName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "Cluster name:\t%s\n", context.Cluster)

	cluster, err := config.LoadClusterConfiguration(config.GetConfigDir(), context.Cluster)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if context.Realm.Name != "" {
		fmt.Fprintf(w, "Astarte Realm:\t%s\n", context.Realm.Name)
		switch {
		case context.Realm.Key != "":
			fmt.Fprintln(w, "Realm Authentication:\tKey")
		case context.Realm.Token != "":
			fmt.Fprintln(w, "Realm Authentication:\tToken")
		default:
			fmt.Fprintln(w, "Realm Authentication:\tNone")
		}
	}
	if cluster.URL != "" {
		fmt.Fprintf(w, "Astarte API URL:\t%s\n", cluster.URL)
	} else {
		fmt.Fprintln(w, "Cluster API URL Type:\tindividual")
	}
	w.Flush()

	return nil
}

func contextsGetRealmKeyF(command *cobra.Command, args []string) error {
	contextName := args[0]
	output, err := command.Flags().GetString("output")
	if err != nil {
		return err
	}

	context, err := config.LoadContextConfiguration(config.GetConfigDir(), contextName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if context.Realm.Key == "" {
		fmt.Fprintf(os.Stderr, "Context %s has no Realm Key associated\n", contextName)
	}

	decoded, err := base64.StdEncoding.DecodeString(context.Realm.Key)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if output != "" {
		// Save to file
		if err := os.WriteFile(output, decoded, 0644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		// Just throw it to stdout
		fmt.Println(string(decoded))
	}

	return nil
}

func contextsCreateF(command *cobra.Command, args []string) error {
	return performContextCreation(args[0], false, command, args)
}

func contextsUpdateF(command *cobra.Command, args []string) error {
	return performContextCreation(args[0], true, command, args)
}

func contextsDeleteF(command *cobra.Command, args []string) error {
	contextName := args[0]

	if _, err := config.LoadContextConfiguration(config.GetConfigDir(), contextName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if ok, err := utils.AskForConfirmation(fmt.Sprintf("Will delete context %s. Are you sure you want to continue?", contextName)); !ok || err != nil {
		return nil
	}

	if err := config.DeleteContextConfiguration(config.GetConfigDir(), contextName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Context %s deleted successfully\n", contextName)
	return nil
}

func performContextCreation(contextName string, isUpdate bool, command *cobra.Command, args []string) error {
	configDir := config.GetConfigDir()
	realmName, err := command.Flags().GetString("realm-name")
	if err != nil {
		return err
	}
	realmPrivateKey, err := command.Flags().GetString("realm-private-key")
	if err != nil {
		return err
	}
	realmToken, err := command.Flags().GetString("realm-token")
	if err != nil {
		return err
	}
	clusterName, err := command.Flags().GetString("cluster")
	if err != nil {
		return err
	}
	activate, err := command.Flags().GetBool("activate")
	if err != nil {
		return err
	}

	// Sanity checks
	switch {
	case (realmPrivateKey != "" || realmToken != "") && realmName == "" && !isUpdate:
		fmt.Fprintln(os.Stderr, "When specifying Realm authentication credentials, you should specify --realm-name")
		os.Exit(1)
	case realmPrivateKey != "" && realmToken != "":
		fmt.Fprintln(os.Stderr, "You should specify only one among --realm-key and --realm-token")
		os.Exit(1)
	}

	if _, err := config.LoadClusterConfiguration(configDir, clusterName); err != nil && clusterName != "" {
		fmt.Fprintf(os.Stderr, "Cluster %s does not exist\n", clusterName)
		os.Exit(1)
	}

	// Check on the context
	context, err := config.LoadContextConfiguration(configDir, contextName)
	switch {
	case err != nil && isUpdate:
		fmt.Fprintf(os.Stderr, "Context %s does not exist. Maybe you meant to use create?\n", contextName)
		os.Exit(1)
	case err == nil && !isUpdate:
		fmt.Fprintf(os.Stderr, "A context named %s already exists. Maybe you meant to use update?\n", contextName)
		os.Exit(1)
	}

	// Go
	if clusterName != "" {
		context.Cluster = clusterName
	}
	if realmName != "" {
		context.Realm.Name = realmName
	}
	if realmPrivateKey != "" {
		contents, err := os.ReadFile(realmPrivateKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		context.Realm.Key = base64.StdEncoding.EncodeToString(contents)
		context.Realm.Token = ""
	}
	if realmToken != "" {
		context.Realm.Token = realmToken
		context.Realm.Key = ""
	}

	// Save
	if err := config.SaveContextConfiguration(configDir, contextName, context, true); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Context %s saved successfully\n", contextName)

	if activate {
		if err := updateCurrentContext(contextName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Printf("Context switched to %s\n", contextName)
		return nil
	}

	return nil
}
