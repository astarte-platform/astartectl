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

package utils

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/astarte-platform/astartectl/utils"

	"github.com/spf13/cobra"
)

var deviceIDCmd = &cobra.Command{
	Use:   "device-id",
	Short: "Various operations on Device IDs",
}

var validateDeviceIDCmd = &cobra.Command{
	Use:   "validate <device_id>",
	Short: "Validates a device ID",
	Long: `Checks whether the provided string is a valid Device ID.

Returns 0 and does not print anything if the ID is valid, returns 1 and prints an error message if it isn't.`,
	Example: `  astartectl utils device-id validate 2TBn-jNESuuHamE2Zo1anA`,
	Args:    cobra.ExactArgs(1),
	RunE:    validateDeviceIDF,
}

var generateRandomDeviceIDCmd = &cobra.Command{
	Use:     "generate-random",
	Short:   "Generates a random Astarte Device ID",
	Long:    `Outputs a valid, random Astarte Device ID.`,
	Example: `  astartectl utils device-id generate-random`,
	RunE:    generateRandomDeviceIDF,
}

var computeDeviceIDFromStringCmd = &cobra.Command{
	Use:   "compute-from-string <namespace_uuid> <string_data>",
	Short: "Computes an Astarte device ID from a UUID and an arbitrary data from a string",
	Long: `Computes a deterministic Astarte device ID using a namespace and arbitrary data from a string.

This leverages UUIDv5 to generate a reproducible ID everytime. A valid UUIDv5 is generated starting from the
supplied data, and it's then encoded into a valid Astarte Device ID. This command is guaranteed to return always
the same ID upon providing the same namespace and data.`,
	Example: `  astartectl utils device-id compute-from-string f79ad91f-c638-4889-ae74-9d001a3b4cf8 myidentifierdata`,
	Args:    cobra.ExactArgs(2),
	RunE:    computeDeviceIDFromStringF,
}

var computeDeviceIDFromBytesCmd = &cobra.Command{
	Use:   "compute-from-bytes <namespace_uuid> <base64_encoded_data>",
	Short: "Computes an Astarte device ID from a UUID and an arbitrary bytearray, encoded in Base64",
	Long: `Computes a deterministic Astarte device ID using a namespace and arbitrary data from a string.

This leverages UUIDv5 to generate a reproducible ID everytime. A valid UUIDv5 is generated starting from the
supplied data, and it's then encoded into a valid Astarte Device ID.

The supplied data must be encoded with Base64 standard encoding. Any input not following this specification will
be rejected.

This command is guaranteed to return always the same ID upon providing the same namespace and data.`,
	Example: `  astartectl utils device-id compute-from-bytes f79ad91f-c638-4889-ae74-9d001a3b4cf8 "bXlpZGVudGlmaWVyZGF0YQ=="`,
	Args:    cobra.ExactArgs(2),
	RunE:    computeDeviceIDFromBytesF,
}

func init() {
	UtilsCmd.AddCommand(deviceIDCmd)

	deviceIDCmd.AddCommand(
		validateDeviceIDCmd,
		generateRandomDeviceIDCmd,
		computeDeviceIDFromStringCmd,
		computeDeviceIDFromBytesCmd,
	)
}

func validateDeviceIDF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	if utils.IsValidAstarteDeviceID(deviceID) {
		fmt.Println("Valid")
		return nil
	}

	fmt.Printf("%s is not a valid Astarte Device ID\n", deviceID)
	os.Exit(1)
	return nil
}

func generateRandomDeviceIDF(command *cobra.Command, args []string) error {
	deviceID, err := utils.GenerateRandomAstarteDeviceID()
	if err != nil {
		return err
	}

	fmt.Println(deviceID)
	return nil
}

func computeDeviceIDFromStringF(command *cobra.Command, args []string) error {
	namespaceUUID := args[0]
	stringData := args[1]
	deviceID, err := utils.GetNamespacedAstarteDeviceID(namespaceUUID, []byte(stringData))
	if err != nil {
		return err
	}

	fmt.Println(deviceID)
	return nil
}

func computeDeviceIDFromBytesF(command *cobra.Command, args []string) error {
	namespaceUUID := args[0]
	bytesData := args[1]
	actualBytes, err := base64.StdEncoding.DecodeString(bytesData)
	if err != nil {
		return err
	}

	deviceID, err := utils.GetNamespacedAstarteDeviceID(namespaceUUID, actualBytes)
	if err != nil {
		return err
	}

	fmt.Println(deviceID)
	return nil
}
