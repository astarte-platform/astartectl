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

	"github.com/spf13/cobra"
)

// MetadataCmd represents the metadata command
var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Interact with Device Metadata",
	Long:  `Perform actions on Astarte Device Metadata.`,
}

var metadataListCmd = &cobra.Command{
	Use:   "list <device_id_or_alias>",
	Short: "List metadata",
	Long: `List all metadata for a given Device.

<device_id_or_alias> can be either a valid Astarte Device ID, or a Device Alias. In most cases,
this is automatically determined - however, you can tweak this behavior by using --force-device-id or
--force-id-type={device-id,alias}.`,
	Example: `  astartectl appengine devices metadata list`,
	Args:    cobra.ExactArgs(1),
	RunE:    metadataListF,
	Aliases: []string{"ls"},
}

var metadataSetCmd = &cobra.Command{
	Use:   "set <device_id_or_alias> <key>=<value>",
	Short: "Set device metadata",
	Long: `Set a value to a metadata key in the given Device, in the form <key>=<value>.
This is effectively an upsert operation: if the key already exist, the new value overwrites the old one.

<device_id_or_alias> can be either a valid Astarte Device ID, or a Device Alias. In most cases,
this is automatically determined - however, you can tweak this behavior by using --force-device-id or
--force-id-type={device-id,alias}.`,
	Example: `  astartectl appengine devices metadata add 2TBn-jNESuuHamE2Zo1anA room=kitchen`,
	Args:    cobra.ExactArgs(2),
	RunE:    metadataSetF,
}

var metadataRemoveCmd = &cobra.Command{
	Use:   "remove <device_id_or_alias> <key>",
	Short: "Remove metadata from a Device",
	Long: `Remove metadata from the given Device, by specifying the key that has to be removed.

<device_id_or_alias> can be either a valid Astarte Device ID, or a Device Alias. In most cases,
this is automatically determined - however, you can tweak this behavior by using --force-device-id or
--force-id-type={device-id,alias}.`,
	Example: `  astartectl appengine devices metadata remove 2TBn-jNESuuHamE2Zo1anA room`,
	Args:    cobra.ExactArgs(2),
	RunE:    metadataRemoveF,
	Aliases: []string{"rm"},
}

func init() {
	devicesCmd.AddCommand(metadataCmd)

	metadataListCmd.Flags().String("force-id-type", "", "When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")
	metadataSetCmd.Flags().String("force-id-type", "", "When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")
	metadataRemoveCmd.Flags().String("force-id-type", "", "When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")

	metadataCmd.AddCommand(
		metadataListCmd,
		metadataSetCmd,
		metadataRemoveCmd,
	)
}

func metadataListF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	forceIDType, err := command.Flags().GetString("force-id-type")
	if err != nil {
		return err
	}
	deviceIdentifierType, err := deviceIdentifierTypeFromFlags(deviceID, forceIDType)
	if err != nil {
		return err
	}

	metadata, err := astarteAPIClient.AppEngine.ListDeviceMetadata(realm, deviceID, deviceIdentifierType)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("%v\n", metadata)
	return nil
}

func metadataSetF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	forceIDType, err := command.Flags().GetString("force-id-type")
	if err != nil {
		return err
	}
	deviceIdentifierType, err := deviceIdentifierTypeFromFlags(deviceID, forceIDType)
	if err != nil {
		return err
	}

	metaKeyAndValue := args[1]
	meta := strings.Split(metaKeyAndValue, "=")
	if len(meta) != 2 {
		fmt.Println("Metadata should be in the form <key>=<value>")
		os.Exit(1)
	}

	err = astarteAPIClient.AppEngine.SetDeviceMetadata(realm, deviceID, deviceIdentifierType, meta[0], meta[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}

func metadataRemoveF(command *cobra.Command, args []string) error {
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

	err = astarteAPIClient.AppEngine.DeleteDeviceMetadata(realm, deviceID, deviceIdentifierType, key)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}
