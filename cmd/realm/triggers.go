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

package realm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// triggersCmd represents the triggers command
var triggersCmd = &cobra.Command{
	Use:     "triggers",
	Short:   "Manage triggers",
	Long:    `List, show, install or delete triggers in your realm.`,
	Aliases: []string{"trigger"},
}

var triggersListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List triggers",
	Long:    `List the name of triggers installed in the realm.`,
	Example: `  astartectl realm-management triggers list`,
	RunE:    triggersListF,
	Aliases: []string{"ls"},
}

var triggersShowCmd = &cobra.Command{
	Use:     "show <trigger_name>",
	Short:   "Show trigger",
	Long:    `Shows a trigger installed in the realm.`,
	Example: `  astartectl realm-management triggers show my_data_trigger`,
	Args:    cobra.ExactArgs(1),
	RunE:    triggersShowF,
}

var triggersInstallCmd = &cobra.Command{
	Use:   "install <trigger_file>",
	Short: "Install trigger",
	Long: `Install the given trigger in the realm.
<trigger_file> must be a path to a JSON file containing a valid Astarte trigger.`,
	Example: `  astartectl realm-management triggers install my_data_trigger.json`,
	Args:    cobra.ExactArgs(1),
	RunE:    triggersInstallF,
}

var triggersDeleteCmd = &cobra.Command{
	Use:     "delete <trigger_name>",
	Short:   "Delete a trigger",
	Long:    `Deletes the specified trigger from the realm.`,
	Example: `  astartectl realm-management triggers delete my_data_trigger`,
	Args:    cobra.ExactArgs(1),
	RunE:    triggersDeleteF,
	Aliases: []string{"del"},
}

var triggersSaveCmd = &cobra.Command{
	Use:   "save [destination-path]",
	Short: "Save triggers to a local folder",
	Long: `Save each trigger in a realm to a local folder. Each trigger will
be saved in a dedicated file whose name will be in the form '<trigger_name>_v<version>.json'.
When no destination path is set, triggers will be saved in the current working directory.
This command does not support the --to-curl flag.`,
	Example: `  astartectl realm-management triggers save`,
	Args:    cobra.MaximumNArgs(1),
	RunE:    triggersSaveF,
}

var triggersSyncCmd = &cobra.Command{
	Use:   "sync <interface_files> [...]",
	Short: "Synchronize triggers",
	Long: `Synchronize triggers in the realm with the given files.
All given files will be parsed, and only new triggers will be installed in the
realm, depending on the realm's state. In order to force triggers update, use --force flag`,
	Example: `  astartectl realm-management triggers sync triggers/*.json`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    triggersSyncF,
}

func init() {

	RealmManagementCmd.AddCommand(triggersCmd)
	triggersSyncCmd.Flags().Bool("force", false, "When set, force triggers update")
	triggersCmd.AddCommand(
		triggersListCmd,
		triggersShowCmd,
		triggersInstallCmd,
		triggersDeleteCmd,
		triggersSaveCmd,
		triggersSyncCmd,
	)
}

func triggersListF(command *cobra.Command, args []string) error {
	realmTriggers, _ := listTriggers(realm)
	fmt.Println(realmTriggers)
	return nil
}

func triggersShowF(command *cobra.Command, args []string) error {

	triggerName := args[0]
	triggerDefinition, _ := getTriggerDefinition(realm, triggerName)
	respJSON, _ := json.MarshalIndent(triggerDefinition, "", "  ")
	fmt.Println(string(respJSON))

	return nil
}

func triggersInstallF(command *cobra.Command, args []string) error {
	triggerFile, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	var triggerBody map[string]interface{}
	err = json.Unmarshal(triggerFile, &triggerBody)
	if err != nil {
		return err
	}

	_ = installTrigger(realm, triggerBody)

	fmt.Println("ok")
	return nil
}

func triggersDeleteF(command *cobra.Command, args []string) error {
	triggerName := args[0]
	deleteTriggerCall, err := astarteAPIClient.DeleteTrigger(realm, triggerName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	utils.MaybeCurlAndExit(deleteTriggerCall, astarteAPIClient)

	deleteTriggerRes, err := deleteTriggerCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, _ = deleteTriggerRes.Parse()

	fmt.Println("ok")
	return nil
}

func triggersSaveF(command *cobra.Command, args []string) error {
	if viper.GetBool("realmmanagement-to-curl") {
		fmt.Println(`'triggers save' does not support the --to-curl option. Use 'triggers list' to get the triggers in your realm, 'triggers versions' to get their versions, and 'triggers show' to get the content of a trigger.`)
		os.Exit(1)
	}

	var targetPath string
	var err error
	if len(args) == 0 {
		targetPath, _ = filepath.Abs(".")
	} else {
		targetPath, err = filepath.Abs(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	// retrieve triggers list
	realmTriggers, err := listTriggers(realm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, name := range realmTriggers {

		triggerDefinition, err := getTriggerDefinition(realm, name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		respJSON, err := json.MarshalIndent(triggerDefinition, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		filename := fmt.Sprintf("/%s/%s.json", targetPath, name)
		outFile, err := os.Create(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer outFile.Close()

		if _, err := outFile.Write(respJSON); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	}
	return nil
}

func triggersSyncF(command *cobra.Command, args []string) error {
	if viper.GetBool("to-curl") {
		fmt.Println(`'triggers sync' does not support the --to-curl option. Install your triggers one by one with 'triggers install'.`)
		os.Exit(1)
	}

	triggerToInstall := []map[string]interface{}{}
	triggerToUpdate := []map[string]interface{}{}

	for _, f := range args {
		triggerFile, err := os.ReadFile(f)
		if err != nil {
			return err
		}

		var astarteTrigger map[string]interface{}
		if err = json.Unmarshal(triggerFile, &astarteTrigger); err != nil {
			return err
		}

		if _, err := getTriggerDefinition(realm, astarteTrigger["name"].(string)); err != nil {
			// The trigger does not exist
			triggerToInstall = append(triggerToInstall, astarteTrigger)
		} else {
			triggerToUpdate = append(triggerToUpdate, astarteTrigger)
		}

		if len(triggerToInstall) == 0 && len(triggerToUpdate) == 0 {
			// All good in the hood
			fmt.Println("Your realm is in sync with the provided triggers files")
			return nil
		}

	}
	// Notify the user about what we're about to do
	list := []string{}

	for _, v := range triggerToInstall {
		list = append(list, v["name"].(string))
	}

	list_existing := []string{}

	for _, v := range triggerToUpdate {
		list_existing = append(list_existing, v["name"].(string))
	}

	// Start syncing.

	//install new triggers
	if len(triggerToInstall) > 0 {

		fmt.Printf("\n")
		fmt.Printf("The following new triggers will be installed: %+q \n", list)
		fmt.Printf("\n")

		if ok, err := utils.AskForConfirmation("Do you want to continue?"); !ok || err != nil {
			fmt.Printf("aborting")
			return nil
		}

		for _, v := range triggerToInstall {
			if err := installTrigger(realm, v); err != nil {
				fmt.Fprintf(os.Stderr, "Could not install trigger %s: %s\n", v["name"].(string), err)
			} else {
				fmt.Printf("trigger %s installed successfully\n", v["name"].(string))
			}
		}

		fmt.Printf("\n")

	}

	if len(triggerToUpdate) > 0 {

		y, err := command.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if y {
			fmt.Printf("The following triggers already exists and WILL be DELETED and RECREATED: %+q \n", list_existing)
			fmt.Printf("\n")
			if ok, err := utils.AskForConfirmation("Do you want to continue?"); !ok || err != nil {
				fmt.Printf("aborting")
				return nil
			}

			for _, v := range triggerToUpdate {
				if err := updateTrigger(realm, v["name"].(string), v); err != nil {
					fmt.Fprintf(os.Stderr, "Could not update trigger %s: %s\n", v["name"].(string), err)
				} else {
					fmt.Printf("trigger %s updated successfully\n", v["name"].(string))
				}
			}
			fmt.Printf("\n")
			fmt.Printf("\n")

		} else {

			// Start syncing.
			fmt.Printf("The following triggers already exists and WILL NOT be updated: %+q \n", list_existing)
			fmt.Printf("\n")

		}
	}

	return nil
}

func installTrigger(realm string, trigger map[string]interface{}) error {
	installTriggerCall, err := astarteAPIClient.InstallTrigger(realm, trigger)
	if err != nil {
		return err
	}

	// When we're here in the context of `triggers sync`, the to-curl flag
	// is always false (`triggers sync` has no `--to-curl` flag)
	// and thus the call will never exit unexpectedly
	utils.MaybeCurlAndExit(installTriggerCall, astarteAPIClient)

	installTriggerRes, err := installTriggerCall.Run(astarteAPIClient)
	if err != nil {
		return err
	}

	_, _ = installTriggerRes.Parse()
	return nil
}

func updateTrigger(realm string, triggername string, newtrig map[string]interface{}) error {

	deleteTriggercall, err := astarteAPIClient.DeleteTrigger(realm, triggername)
	if err != nil {
		return err
	}
	utils.MaybeCurlAndExit(deleteTriggercall, astarteAPIClient)

	_, err = deleteTriggercall.Run(astarteAPIClient)
	if err != nil {
		return err
	}

	updateTriggerCall, err := astarteAPIClient.InstallTrigger(realm, newtrig)
	if err != nil {
		return err
	}

	// When we're here in the context of `triggers sync`, the to-curl flag
	// is always false (`triggers sync` has no `--to-curl` flag)
	// and thus the call will never exit unexpectedly
	utils.MaybeCurlAndExit(updateTriggerCall, astarteAPIClient)

	updateTriggerRes, err := updateTriggerCall.Run(astarteAPIClient)
	if err != nil {
		return err
	}

	_, _ = updateTriggerRes.Parse()
	return nil
}

func listTriggers(realm string) ([]string, error) {
	listTriggersCall, err := astarteAPIClient.ListTriggers(realm)
	if err != nil {
		return []string{}, err
	}

	utils.MaybeCurlAndExit(listTriggersCall, astarteAPIClient)

	listTriggersRes, err := listTriggersCall.Run(astarteAPIClient)
	if err != nil {
		return []string{}, err
	}
	rawlistTriggers, err := listTriggersRes.Parse()
	if err != nil {
		return []string{}, err
	}
	return rawlistTriggers.([]string), nil
}

func getTriggerDefinition(realm, triggerName string) (map[string]interface{}, error) {
	getTriggerCall, err := astarteAPIClient.GetTrigger(realm, triggerName)
	if err != nil {
		return nil, err
	}

	// When we're here in the context of `interfaces sync`, the to-curl flag
	// is always false (`interfaces sync` has no `--to-curl` flag)
	// and thus the call will never exit unexpectedly
	utils.MaybeCurlAndExit(getTriggerCall, astarteAPIClient)

	getTriggerRes, err := getTriggerCall.Run(astarteAPIClient)
	if err != nil {
		return nil, err
	}
	rawTRigger, err := getTriggerRes.Parse()
	if err != nil {
		return nil, err
	}
	triggerDefinition, _ := rawTRigger.(map[string]interface{})
	return triggerDefinition, nil
}
