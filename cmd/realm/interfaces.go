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
	"strconv"

	"github.com/astarte-platform/astarte-go/interfaces"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// interfacesCmd represents the interfaces command
var interfacesCmd = &cobra.Command{
	Use:     "interfaces",
	Short:   "Manage interfaces",
	Long:    `List, show, install or update interfaces in your realm.`,
	Aliases: []string{"interface"},
}

var interfacesListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List interfaces",
	Long:    `List the name of the interfaces installed in the realm.`,
	Example: `  astartectl realm-management interfaces list`,
	RunE:    interfacesListF,
	Aliases: []string{"ls"},
}

var interfacesVersionsCmd = &cobra.Command{
	Use:     "versions <interface_name>",
	Short:   "List major versions of an interface",
	Long:    `List the major versions of an interface installed in the realm.`,
	Example: `  astartectl realm-management interfaces versions com.my.Interface`,
	Args:    cobra.ExactArgs(1),
	RunE:    interfacesVersionsF,
}

var interfacesShowCmd = &cobra.Command{
	Use:     "show <interface_name> <interface_major>",
	Short:   "Show interface",
	Long:    `Show the given major version of the interface installed in the realm.`,
	Example: `  astartectl realm-management interfaces show com.my.Interface 0`,
	Args:    cobra.ExactArgs(2),
	RunE:    interfacesShowF,
}

var interfacesInstallCmd = &cobra.Command{
	Use:   "install <interface_file>",
	Short: "Install interface",
	Long: `Install the given interface in the realm.
<interface_file> must be a path to a JSON file containing a valid Astarte interface.`,
	Example: `  astartectl realm-management interfaces install com.my.Interface.json`,
	Args:    cobra.ExactArgs(1),
	RunE:    interfacesInstallF,
}

var interfacesDeleteCmd = &cobra.Command{
	Use:   "delete <interface_name>",
	Short: "Delete a draft interface",
	Long: `Deletes the specified interface from the realm.
Only draft interfaces for which no devices has sent data to can be removed - as such,
only Major Version 0 of <interface_name> will be deleted, if existing.
Non-draft interfaces should be removed manually or by your system administrator.`,
	Example: `  astartectl realm-management interfaces delete com.my.Interface`,
	Args:    cobra.ExactArgs(1),
	RunE:    interfacesDeleteF,
	Aliases: []string{"del"},
}

var interfacesUpdateCmd = &cobra.Command{
	Use:   "update <interface_file>",
	Short: "Update interface",
	Long: `Update the given interface in the realm.
<interface_file> must be a path to a JSON file containing a valid Astarte interface.

The name and major version of the interface are read from the interface file.`,
	Example: `  astartectl realm-management interfaces update com.my.Interface.json`,
	Args:    cobra.ExactArgs(1),
	RunE:    interfacesUpdateF,
}

var interfacesSyncCmd = &cobra.Command{
	Use:   "sync <interface_files> [...]",
	Short: "Synchronize interfaces",
	Long: `Synchronize interfaces in the realm with the given files.
All given files will be parsed, and interfaces will be either updated or installed in the
realm, depending on the realm's state.`,
	Example: `  astartectl realm-management interfaces sync interfaces/*.json`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    interfacesSyncF,
}

var interfacesSaveCmd = &cobra.Command{
	Use:   "save [destination-path]",
	Short: "Save interfaces to a local folder",
	Long: `Save each interface in a realm to a local folder. Each interface will
be saved in a dedicated file whose name will be in the form '<interface_name>_v<version>.json'.
When no destination path is set, interfaces will be saved in the current working directory.
This command does not support the --to-curl flag.`,
	Example: `  astartectl realm-management interfaces save`,
	Args:    cobra.MaximumNArgs(1),
	RunE:    interfacesSaveF,
}

func init() {
	RealmManagementCmd.AddCommand(interfacesCmd)

	interfacesSyncCmd.PersistentFlags().BoolP("non-interactive", "y", false, "Non-interactive mode. Will answer yes by default to all questions.")

	interfacesCmd.AddCommand(
		interfacesListCmd,
		interfacesVersionsCmd,
		interfacesShowCmd,
		interfacesInstallCmd,
		interfacesDeleteCmd,
		interfacesUpdateCmd,
		interfacesSyncCmd,
		interfacesSaveCmd,
	)
}

func interfacesListF(command *cobra.Command, args []string) error {
	realmInterfaces, err := listInterfaces(realm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(realmInterfaces)
	return nil
}

func interfacesVersionsF(command *cobra.Command, args []string) error {
	interfaceName := args[0]

	interfaceVersions, err := interfaceVersions(interfaceName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(interfaceVersions)
	return nil
}

func interfacesShowF(command *cobra.Command, args []string) error {
	interfaceName := args[0]
	interfaceMajorString := args[1]
	interfaceMajor, err := strconv.Atoi(interfaceMajorString)
	if err != nil {
		return err
	}
	interfaceDefinition, err := getInterfaceDefinition(realm, interfaceName, interfaceMajor)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	respJSON, err := json.MarshalIndent(interfaceDefinition, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(respJSON))
	return nil
}

func interfacesInstallF(command *cobra.Command, args []string) error {
	interfaceFile, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	var interfaceBody interfaces.AstarteInterface
	if err = json.Unmarshal(interfaceFile, &interfaceBody); err != nil {
		return err
	}

	if err = installInterface(realm, interfaceBody); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}

func interfacesDeleteF(command *cobra.Command, args []string) error {
	interfaceName := args[0]
	interfaceMajor := 0

	deleteInterfaceCall, err := astarteAPIClient.DeleteInterface(realm, interfaceName, interfaceMajor)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	utils.MaybeCurlAndExit(deleteInterfaceCall, astarteAPIClient)

	deleteInterfaceRes, err := deleteInterfaceCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, _ = deleteInterfaceRes.Parse()

	fmt.Println("ok")
	return nil
}

func interfacesUpdateF(command *cobra.Command, args []string) error {
	interfaceFile, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	var astarteInterface interfaces.AstarteInterface
	if err = json.Unmarshal(interfaceFile, &astarteInterface); err != nil {
		return err
	}

	if err := updateInterface(realm, astarteInterface.Name, astarteInterface.MajorVersion, astarteInterface); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}

func interfacesSyncF(command *cobra.Command, args []string) error {
	// `interface sync` is unnatural btw
	if viper.GetBool("to-curl") {
		fmt.Println(`'interfaces sync' does not support the --to-curl option.
Install or update your interfaces one by one with 'interfaces install' or 'interface update'.`)
		os.Exit(1)
	}

	interfacesToInstall := []interfaces.AstarteInterface{}
	interfacesToUpdate := []interfaces.AstarteInterface{}

	for _, f := range args {
		interfaceFile, err := os.ReadFile(f)
		if err != nil {
			return err
		}

		var astarteInterface interfaces.AstarteInterface
		if err = json.Unmarshal(interfaceFile, &astarteInterface); err != nil {
			return err
		}

		if interfaceDefinition, err := getInterfaceDefinition(realm, astarteInterface.Name, astarteInterface.MajorVersion); err != nil {
			// The interface does not exist
			interfacesToInstall = append(interfacesToInstall, astarteInterface)
		} else {
			if interfaceDefinition.MinorVersion < astarteInterface.MinorVersion {
				interfacesToUpdate = append(interfacesToUpdate, astarteInterface)
			} else if interfaceDefinition.MinorVersion > astarteInterface.MinorVersion {
				// Notify that the realm has a more recent revision
				fmt.Fprintf(os.Stderr, "warn: Interface %s has version %d.%d in the realm and %d.%d in the local file", interfaceDefinition.Name,
					interfaceDefinition.MajorVersion, interfaceDefinition.MinorVersion, astarteInterface.MajorVersion, astarteInterface.MinorVersion)
			}
		}
	}

	if len(interfacesToInstall) == 0 && len(interfacesToUpdate) == 0 {
		// All good in the hood
		fmt.Println("Your realm is in sync with the provided interface files")
		return nil
	}

	// Notify the user about what we're about to do
	fmt.Println("The following actions will be taken:")
	fmt.Println()
	for _, v := range interfacesToInstall {
		fmt.Printf("Will install interface %s version %d.%d\n", v.Name, v.MajorVersion, v.MinorVersion)
	}
	for _, v := range interfacesToUpdate {
		fmt.Printf("Will update interface %s to version %d.%d\n", v.Name, v.MajorVersion, v.MinorVersion)
	}
	fmt.Println()

	y, err := command.Flags().GetBool("non-interactive")
	if err != nil {
		return err
	}
	if !y {
		if ok, err := utils.AskForConfirmation("Do you want to continue?"); !ok || err != nil {
			return nil
		}
	}

	// Start syncing.
	for _, v := range interfacesToInstall {
		if err := installInterface(realm, v); err != nil {
			fmt.Fprintf(os.Stderr, "Could not install interface %s: %s\n", v.Name, err)
		} else {
			fmt.Printf("Interface %s installed successfully\n", v.Name)
		}
	}
	for _, v := range interfacesToUpdate {
		if err := updateInterface(realm, v.Name, v.MajorVersion, v); err != nil {
			fmt.Fprintf(os.Stderr, "Could not update interface %s: %s\n", v.Name, err)
		} else {
			fmt.Printf("Interface %s updated successfully to version %d.%d\n", v.Name, v.MajorVersion, v.MinorVersion)
		}
	}

	return nil
}

func getInterfaceDefinition(realm, interfaceName string, interfaceMajor int) (interfaces.AstarteInterface, error) {
	getInterfaceCall, err := astarteAPIClient.GetInterface(realm, interfaceName, interfaceMajor)
	if err != nil {
		return interfaces.AstarteInterface{}, err
	}

	// When we're here in the context of `interfaces sync`, the to-curl flag
	// is always false (`interfaces sync` has no `--to-curl` flag)
	// and thus the call will never exit unexpectedly
	utils.MaybeCurlAndExit(getInterfaceCall, astarteAPIClient)

	getInterfaceRes, err := getInterfaceCall.Run(astarteAPIClient)
	if err != nil {
		return interfaces.AstarteInterface{}, err
	}
	rawInterface, err := getInterfaceRes.Parse()
	if err != nil {
		return interfaces.AstarteInterface{}, err
	}
	interfaceDefinition, _ := rawInterface.(interfaces.AstarteInterface)
	return interfaceDefinition, nil
}

func installInterface(realm string, iface interfaces.AstarteInterface) error {
	installInterfaceCall, err := astarteAPIClient.InstallInterface(realm, iface, true)
	if err != nil {
		return err
	}

	// When we're here in the context of `interfaces sync`, the to-curl flag
	// is always false (`interfaces sync` has no `--to-curl` flag)
	// and thus the call will never exit unexpectedly
	utils.MaybeCurlAndExit(installInterfaceCall, astarteAPIClient)

	installInterfaceRes, err := installInterfaceCall.Run(astarteAPIClient)
	if err != nil {
		return err
	}

	_, _ = installInterfaceRes.Parse()
	return nil
}

func updateInterface(realm string, interfaceName string, interfaceMajor int, newInterface interfaces.AstarteInterface) error {
	updateInterfaceCall, err := astarteAPIClient.UpdateInterface(realm, interfaceName, interfaceMajor, newInterface, true)
	if err != nil {
		return err
	}

	// When we're here in the context of `interfaces sync`, the to-curl flag
	// is always false (`interfaces sync` has no `--to-curl` flag)
	// and thus the call will never exit unexpectedly
	utils.MaybeCurlAndExit(updateInterfaceCall, astarteAPIClient)

	updateInterfaceRes, err := updateInterfaceCall.Run(astarteAPIClient)
	if err != nil {
		return err
	}

	_, _ = updateInterfaceRes.Parse()
	return nil
}

func interfacesSaveF(command *cobra.Command, args []string) error {
	if viper.GetBool("realmmanagement-to-curl") {
		fmt.Println(`'interfaces save' does not support the --to-curl option. Use 'interfaces list' to get the interfaces in your realm, 'interfaces versions' to get their versions, and 'interfaces show' to get the content of an interface.`)
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

	// retrieve interfaces list
	realmInterfaces, err := listInterfaces(realm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ifaceNameAndVersions := map[string][]int{}

	// and the versions for each interface
	for _, ifaceName := range realmInterfaces {
		interfaceVersions, err := interfaceVersions(ifaceName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		ifaceNameAndVersions[ifaceName] = interfaceVersions
	}

	for name, versions := range ifaceNameAndVersions {
		for _, v := range versions {
			interfaceDefinition, err := getInterfaceDefinition(realm, name, v)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			respJSON, err := json.MarshalIndent(interfaceDefinition, "", "  ")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			filename := fmt.Sprintf("/%s/%s_v%d.json", targetPath, name, v)
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
	}
	return nil
}

func listInterfaces(realm string) ([]string, error) {
	listInterfacesCall, err := astarteAPIClient.ListInterfaces(realm)
	if err != nil {
		return []string{}, err
	}

	utils.MaybeCurlAndExit(listInterfacesCall, astarteAPIClient)

	listInterfacesRes, err := listInterfacesCall.Run(astarteAPIClient)
	if err != nil {
		return []string{}, err
	}
	rawListInterfaces, err := listInterfacesRes.Parse()
	if err != nil {
		return []string{}, err
	}
	return rawListInterfaces.([]string), nil
}

func interfaceVersions(interfaceName string) ([]int, error) {
	interfaceVersionsCall, err := astarteAPIClient.ListInterfaceMajorVersions(realm, interfaceName)
	if err != nil {
		return []int{}, err
	}

	utils.MaybeCurlAndExit(interfaceVersionsCall, astarteAPIClient)

	interfaceVersionsRes, err := interfaceVersionsCall.Run(astarteAPIClient)
	if err != nil {
		return []int{}, err
	}
	rawInterfaceVersions, err := interfaceVersionsRes.Parse()
	if err != nil {
		return []int{}, err
	}
	return rawInterfaceVersions.([]int), nil
}
