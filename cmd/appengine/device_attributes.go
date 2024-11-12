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

package appengine

import (
	"fmt"
	"os"
	"strings"

	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
)

// AttributesCmd represents the attributes command
var attributesCmd = &cobra.Command{
	Use:   "attributes",
	Short: "Interact with Device Attributes",
	Long:  `Perform actions on Astarte Device Attributes.`,
}

var attributesListCmd = &cobra.Command{
	Use:   "list <device_id_or_alias>",
	Short: "List attributes",
	Long: `List all attributes for a given Device.

<device_id_or_alias> can be either a valid Astarte Device ID, or a Device Alias. In most cases,
this is automatically determined - however, you can tweak this behavior by using --force-device-id or
--force-id-type={device-id,alias}.`,
	Example: `  astartectl appengine devices attributes list`,
	Args:    cobra.ExactArgs(1),
	RunE:    attributesListF,
	Aliases: []string{"ls"},
}

var attributeSetCmd = &cobra.Command{
	Use:   "set <device_id_or_alias> <key>=<value>",
	Short: "Set device attribute",
	Long: `Set a value to an attribute key in the given Device, in the form <key>=<value>.
This is effectively an upsert operation: if the key already exist, the new value overwrites the old one.

<device_id_or_alias> can be either a valid Astarte Device ID, or a Device Alias. In most cases,
this is automatically determined - however, you can tweak this behavior by using --force-device-id or
--force-id-type={device-id,alias}.`,
	Example: `  astartectl appengine devices attributes add 2TBn-jNESuuHamE2Zo1anA room=kitchen`,
	Args:    cobra.ExactArgs(2),
	RunE:    attributeSetF,
}

var attributeRemoveCmd = &cobra.Command{
	Use:   "remove <device_id_or_alias> <key>",
	Short: "Remove attribute from a Device",
	Long: `Remove attribute from the given Device, by specifying the key that has to be removed.

<device_id_or_alias> can be either a valid Astarte Device ID, or a Device Alias. In most cases,
this is automatically determined - however, you can tweak this behavior by using --force-device-id or
--force-id-type={device-id,alias}.`,
	Example: `  astartectl appengine devices attributes remove 2TBn-jNESuuHamE2Zo1anA room`,
	Args:    cobra.ExactArgs(2),
	RunE:    attributeRemoveF,
	Aliases: []string{"rm"},
}

func init() {
	devicesCmd.AddCommand(attributesCmd)

	attributesListCmd.Flags().String("force-id-type", "", "When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")
	attributeSetCmd.Flags().String("force-id-type", "", "When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")
	attributeRemoveCmd.Flags().String("force-id-type", "", "When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")

	attributesCmd.AddCommand(
		attributesListCmd,
		attributeSetCmd,
		attributeRemoveCmd,
	)
}

func attributesListF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	forceIDType, err := command.Flags().GetString("force-id-type")
	if err != nil {
		return err
	}
	deviceIdentifierType, err := deviceIdentifierTypeFromFlags(deviceID, forceIDType)
	if err != nil {
		return err
	}

	attributesCall, err := astarteAPIClient.ListDeviceAttributes(realm, deviceID, deviceIdentifierType)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	utils.MaybeCurlAndExit(attributesCall, astarteAPIClient)

	attributesRes, err := attributesCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	attributes, _ := attributesRes.Parse()
	attributesMap, _ := attributes.(map[string]string)

	for k, v := range attributesMap {
		fmt.Printf("%v: %v\n", k, v)
	}

	return nil
}

func attributeSetF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	forceIDType, err := command.Flags().GetString("force-id-type")
	if err != nil {
		return err
	}
	deviceIdentifierType, err := deviceIdentifierTypeFromFlags(deviceID, forceIDType)
	if err != nil {
		return err
	}

	attributeKeyAndValue := args[1]
	attr := strings.Split(attributeKeyAndValue, "=")
	if len(attr) != 2 {
		fmt.Println("Attributes should be in the form <key>=<value>")
		os.Exit(1)
	}

	setAttributeCall, err := astarteAPIClient.SetDeviceAttribute(realm, deviceID, deviceIdentifierType, attr[0], attr[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	utils.MaybeCurlAndExit(setAttributeCall, astarteAPIClient)

	setAttributeRes, err := setAttributeCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, _ = setAttributeRes.Parse()

	fmt.Println("ok")
	return nil
}

func attributeRemoveF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	forceIDType, err := command.Flags().GetString("force-id-type")
	if err != nil {
		return err
	}
	deviceIdentifierType, err := deviceIdentifierTypeFromFlags(deviceID, forceIDType)
	if err != nil {
		return err
	}

	key := args[1]

	deleteAttributeCall, err := astarteAPIClient.DeleteDeviceAttribute(realm, deviceID, deviceIdentifierType, key)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	utils.MaybeCurlAndExit(deleteAttributeCall, astarteAPIClient)

	deleteAttributeRes, err := deleteAttributeCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, _ = deleteAttributeRes.Parse()

	fmt.Println("ok")
	return nil
}
