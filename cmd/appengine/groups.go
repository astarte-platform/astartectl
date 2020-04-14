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

package appengine

import (
	"fmt"
	"os"
	"strings"

	"github.com/astarte-platform/astarte-go/client"
	"github.com/spf13/cobra"
)

// groupsCmd represents the groups command
var groupsCmd = &cobra.Command{
	Use:     "groups",
	Short:   "Manage groups",
	Long:    `List, show, create or update groups in your realm.`,
	Aliases: []string{"group"},
}

var groupsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List groups",
	Long:    `List the name of the groups installed in the realm.`,
	Example: `  astartectl appengine groups list`,
	RunE:    groupsListF,
}

var groupsCreateCmd = &cobra.Command{
	Use:   "create <group_name> <device_list>",
	Short: "Create a group",
	Long: `Create a group in the realm.
<device_list> must be a comma separated list of Device identifiers (i.e. a Device ID or an alias).
All devices must already be registered in the realm.`,
	Example: `  astartectl appengine groups create mygroup dI2dZrblSbObnAazrduIDw,r0mDcECmSa2exhGCs7D38A`,
	Args:    cobra.ExactArgs(2),
	RunE:    groupsCreateF,
}

var groupsDevicesCmd = &cobra.Command{
	Use:     "devices",
	Short:   "Manage devices in a group",
	Long:    `List, add, or remove devices in a group.`,
	Aliases: []string{"device"},
}

var groupsDevicesListCmd = &cobra.Command{
	Use:     "list <group_name>",
	Short:   "List devices in a group",
	Long:    `List devices in a group`,
	Example: `  astartectl appengine groups devices list mygroup`,
	Args:    cobra.ExactArgs(1),
	RunE:    groupsDevicesListF,
}

var groupsDevicesAddCmd = &cobra.Command{
	Use:     "add <group_name> <device_id_or_alias>",
	Short:   "Add a device to a group",
	Long:    `Add a device to a group`,
	Example: `  astartectl appengine groups devices add mygroup 7O1hqtg0TSyKpNXr_AqEJA`,
	Args:    cobra.ExactArgs(2),
	RunE:    groupsDevicesAddF,
}

var groupsDevicesRemoveCmd = &cobra.Command{
	Use:     "remove <group_name> <device_id_or_alias>",
	Short:   "Remove a device from a group",
	Long:    `Remove a device from a group`,
	Example: `  astartectl appengine groups devices remove mygroup y3QgB6BAST2BGGK8GtNSmQ`,
	Args:    cobra.ExactArgs(2),
	RunE:    groupsDevicesRemoveF,
}

func init() {
	groupsCreateCmd.Flags().String("force-id-type", "",
		"When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")

	groupsDevicesAddCmd.Flags().String("force-id-type", "",
		"When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")

	groupsDevicesRemoveCmd.Flags().String("force-id-type", "",
		"When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")

	groupsDevicesCmd.AddCommand(
		groupsDevicesListCmd,
		groupsDevicesAddCmd,
		groupsDevicesRemoveCmd,
	)

	groupsCmd.AddCommand(
		groupsListCmd,
		groupsCreateCmd,
		groupsDevicesCmd,
	)

	AppEngineCmd.AddCommand(groupsCmd)
}

func groupsListF(command *cobra.Command, args []string) error {
	groupsList, err := astarteAPIClient.AppEngine.ListGroups(realm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(groupsList)
	return nil
}

func groupsCreateF(command *cobra.Command, args []string) error {
	groupName := args[0]
	devices := args[1]

	forceIDType, err := command.Flags().GetString("force-id-type")
	if err != nil {
		return err
	}

	deviceIdentifiers := strings.Split(devices, ",")
	var deviceIdentifiersType client.DeviceIdentifierType
	// Check all identifiers, we'll use the last deviceIdentifierType set in the loop
	for _, deviceIdentifier := range deviceIdentifiers {
		deviceIdentifiersType, err = deviceIdentifierTypeFromFlags(deviceIdentifier, forceIDType)
		if err != nil {
			return err
		}
	}
	err = astarteAPIClient.AppEngine.CreateGroup(realm, groupName, deviceIdentifiers, deviceIdentifiersType)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}

func groupsDevicesListF(command *cobra.Command, args []string) error {
	groupName := args[0]

	deviceList, err := astarteAPIClient.AppEngine.ListGroupDevices(realm, groupName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(deviceList)
	return nil
}

func groupsDevicesAddF(command *cobra.Command, args []string) error {
	groupName := args[0]
	deviceIdentifier := args[1]

	forceIDType, err := command.Flags().GetString("force-id-type")
	if err != nil {
		return err
	}

	deviceIdentifierType, err := deviceIdentifierTypeFromFlags(deviceIdentifier, forceIDType)
	if err != nil {
		return err
	}

	err = astarteAPIClient.AppEngine.AddDeviceToGroup(realm, groupName, deviceIdentifier, deviceIdentifierType)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}

func groupsDevicesRemoveF(command *cobra.Command, args []string) error {
	groupName := args[0]
	deviceIdentifier := args[1]

	forceIDType, err := command.Flags().GetString("force-id-type")
	if err != nil {
		return err
	}

	deviceIdentifierType, err := deviceIdentifierTypeFromFlags(deviceIdentifier, forceIDType)
	if err != nil {
		return err
	}

	err = astarteAPIClient.AppEngine.RemoveDeviceFromGroup(realm, groupName, deviceIdentifier, deviceIdentifierType)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}
