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
	"io/ioutil"
	"os"
	"strconv"

	"github.com/astarte-platform/astarte-go/interfaces"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
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

func init() {
	RealmManagementCmd.AddCommand(interfacesCmd)

	interfacesCmd.AddCommand(
		interfacesListCmd,
		interfacesVersionsCmd,
		interfacesShowCmd,
		interfacesInstallCmd,
		interfacesDeleteCmd,
		interfacesUpdateCmd,
		interfacesSyncCmd,
	)
}

func interfacesListF(command *cobra.Command, args []string) error {
	realmInterfaces, err := astarteAPIClient.RealmManagement.ListInterfaces(realm)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(realmInterfaces)
	return nil
}

func interfacesVersionsF(command *cobra.Command, args []string) error {
	interfaceName := args[0]
	interfaceVersions, err := astarteAPIClient.RealmManagement.ListInterfaceMajorVersions(realm, interfaceName)
	if err != nil {
		fmt.Println(err)
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

	interfaceDefinition, err := astarteAPIClient.RealmManagement.GetInterface(realm, interfaceName, interfaceMajor)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	respJSON, err := json.MarshalIndent(interfaceDefinition, "", "  ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(respJSON))
	return nil
}

func interfacesInstallF(command *cobra.Command, args []string) error {
	interfaceFile, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	var interfaceBody interfaces.AstarteInterface
	if err = json.Unmarshal(interfaceFile, &interfaceBody); err != nil {
		return err
	}

	if err = astarteAPIClient.RealmManagement.InstallInterface(realm, interfaceBody); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}

func interfacesDeleteF(command *cobra.Command, args []string) error {
	interfaceName := args[0]
	interfaceMajor := 0

	if err := astarteAPIClient.RealmManagement.DeleteInterface(realm, interfaceName, interfaceMajor); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}

func interfacesUpdateF(command *cobra.Command, args []string) error {
	interfaceFile, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	var astarteInterface interfaces.AstarteInterface
	if err = json.Unmarshal(interfaceFile, &astarteInterface); err != nil {
		return err
	}

	if err = astarteAPIClient.RealmManagement.UpdateInterface(realm, astarteInterface.Name, astarteInterface.MajorVersion,
		astarteInterface); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}

func interfacesSyncF(command *cobra.Command, args []string) error {
	interfacesToInstall := []interfaces.AstarteInterface{}
	interfacesToUpdate := []interfaces.AstarteInterface{}

	for _, f := range args {
		interfaceFile, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}

		var astarteInterface interfaces.AstarteInterface
		if err = json.Unmarshal(interfaceFile, &astarteInterface); err != nil {
			return err
		}

		if interfaceDefinition, err := astarteAPIClient.RealmManagement.GetInterface(realm, astarteInterface.Name, astarteInterface.MajorVersion); err != nil {
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
	if ok, err := utils.AskForConfirmation("Do you want to continue?"); !ok || err != nil {
		return nil
	}

	// Start syncing.
	for _, v := range interfacesToInstall {
		if err := astarteAPIClient.RealmManagement.InstallInterface(realm, v); err != nil {
			fmt.Fprintf(os.Stderr, "Could not install interface %s: %s\n", v.Name, err)
		} else {
			fmt.Printf("Interface %s installed successfully\n", v.Name)
		}
	}
	for _, v := range interfacesToUpdate {
		if err := astarteAPIClient.RealmManagement.UpdateInterface(realm, v.Name, v.MajorVersion, v); err != nil {
			fmt.Fprintf(os.Stderr, "Could not update interface %s: %s\n", v.Name, err)
		} else {
			fmt.Printf("Interface %s updated successfully to version %d.%d\n", v.Name, v.MajorVersion, v.MinorVersion)
		}
	}

	return nil
}
