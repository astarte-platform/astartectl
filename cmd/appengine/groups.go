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
	groupsListCall, err := astarteAPIClient.ListGroups(realm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	groupsListRes, err := groupsListCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	groupsList, _ := groupsListRes.Parse()

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

	// turn device aliases in device IDs, if needed
	if deviceIdentifiersType == client.AstarteDeviceAlias {
		for i := 0; i < len(deviceIdentifiers); i++ {
			deviceIdentifiers[i], err = getDeviceIDfromAlias(deviceIdentifiers[i])
			if err != nil {
				return err
			}
		}
	}

	createGroupCall, err := astarteAPIClient.CreateGroup(realm, groupName, deviceIdentifiers)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	createGroupRes, err := createGroupCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_, _ = createGroupRes.Parse()

	fmt.Println("ok")
	return nil
}

func groupsDevicesListF(command *cobra.Command, args []string) error {
	groupName := args[0]

	deviceListPaginator, err := astarteAPIClient.ListGroupDevices(realm, groupName, 100, client.DeviceIDFormat)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	deviceList := []string{}
	for deviceListPaginator.HasNextPage() {
		deviceListCall, err := deviceListPaginator.GetNextPage()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		deviceListRes, err := deviceListCall.Run(astarteAPIClient)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		rawDevices, _ := deviceListRes.Parse()
		devices, _ := rawDevices.([]string)
		for _, v := range devices {
			deviceList = append(deviceList, v)
		}
	}

	fmt.Println(deviceList)
	return nil
}

func groupsDevicesAddF(command *cobra.Command, args []string) error {
	groupName := args[0]
	deviceIdentifier, err := getDeviceIDfromArgs(command, args)
	if err != nil {
		return err
	}

	addDeviceCall, err := astarteAPIClient.AddDeviceToGroup(realm, groupName, deviceIdentifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	addDeviceRes, err := addDeviceCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, _ = addDeviceRes.Parse()

	fmt.Println("ok")
	return nil
}

func groupsDevicesRemoveF(command *cobra.Command, args []string) error {
	groupName := args[0]
	deviceIdentifier, err := getDeviceIDfromArgs(command, args)
	if err != nil {
		return err
	}

	removeDeviceCall, err := astarteAPIClient.RemoveDeviceFromGroup(realm, groupName, deviceIdentifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	removeDeviceRes, err := removeDeviceCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, _ = removeDeviceRes.Parse()

	fmt.Println("ok")
	return nil
}

func getDeviceIDfromArgs(cmd *cobra.Command, args []string) (string, error) {
	deviceIdentifier := args[1]

	forceIDType, err := cmd.Flags().GetString("force-id-type")
	if err != nil {
		return "", err
	}

	deviceIdentifierType, err := deviceIdentifierTypeFromFlags(deviceIdentifier, forceIDType)
	if err != nil {
		return "", err
	}

	if deviceIdentifierType == client.AstarteDeviceAlias {
		deviceIdentifier, err = getDeviceIDfromAlias(deviceIdentifier)
		if err != nil {
			return "", err
		}
	}

	return deviceIdentifier, nil
}

func getDeviceIDfromAlias(alias string) (string, error) {
	getDeviceIDCall, err := astarteAPIClient.GetDeviceIDFromAlias(realm, alias)
	if err != nil {
		return "", fmt.Errorf("Could not resolve the alias %s to an Astarte Device ID, error %w", alias, err)
	}
	getDeviceIDRes, err := getDeviceIDCall.Run(astarteAPIClient)
	if err != nil {
		return "", fmt.Errorf("Could not resolve the alias %s to an Astarte Device ID, error %w", alias, err)
	}
	rawDeviceID, _ := getDeviceIDRes.Parse()
	deviceID, _ := rawDeviceID.(string)
	return deviceID, nil
}
