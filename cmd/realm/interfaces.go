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

	"github.com/spf13/cobra"
)

// interfacesCmd represents the interfaces command
var interfacesCmd = &cobra.Command{
	Use:   "interfaces",
	Short: "Manage interfaces",
	Long:  `List, show, install or update interfaces in your realm.`,
}

var interfacesListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List interfaces",
	Long:    `List the name of the interfaces installed in the realm.`,
	Example: `  astartectl realm-management interfaces list`,
	RunE:    interfacesListF,
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

func init() {
	RealmManagementCmd.AddCommand(interfacesCmd)

	interfacesCmd.AddCommand(
		interfacesListCmd,
		interfacesVersionsCmd,
		interfacesShowCmd,
		interfacesInstallCmd,
		interfacesDeleteCmd,
		interfacesUpdateCmd,
	)
}

func interfacesListF(command *cobra.Command, args []string) error {
	realmInterfaces, err := astarteAPIClient.RealmManagement.ListInterfaces(realm, realmManagementJwt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(realmInterfaces)
	return nil
}

func interfacesVersionsF(command *cobra.Command, args []string) error {
	interfaceName := args[0]
	interfaceVersions, err := astarteAPIClient.RealmManagement.ListInterfaceMajorVersions(realm, interfaceName, realmManagementJwt)
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

	interfaceDefinition, err := astarteAPIClient.RealmManagement.GetInterface(realm, interfaceName, interfaceMajor, realmManagementJwt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	respJSON, _ := json.MarshalIndent(interfaceDefinition, "", "  ")
	fmt.Println(string(respJSON))
	return nil
}

func interfacesInstallF(command *cobra.Command, args []string) error {
	interfaceFile, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	var interfaceBody map[string]interface{}
	err = json.Unmarshal(interfaceFile, &interfaceBody)
	if err != nil {
		return err
	}

	err = astarteAPIClient.RealmManagement.InstallInterface(realm, interfaceBody, realmManagementJwt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}

func interfacesDeleteF(command *cobra.Command, args []string) error {
	interfaceName := args[0]
	interfaceMajor := 0

	err := astarteAPIClient.RealmManagement.DeleteInterface(realm, interfaceName, interfaceMajor, realmManagementJwt)
	if err != nil {
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

	var interfaceBody map[string]interface{}
	err = json.Unmarshal(interfaceFile, &interfaceBody)
	if err != nil {
		return err
	}

	interfaceName := fmt.Sprintf("%v", interfaceBody["interface_name"])
	interfaceMajorString := fmt.Sprintf("%v", interfaceBody["version_major"])
	interfaceMajor, err := strconv.Atoi(interfaceMajorString)
	if err != nil {
		return err
	}

	err = astarteAPIClient.RealmManagement.UpdateInterface(realm, interfaceName, interfaceMajor, interfaceBody, realmManagementJwt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}
