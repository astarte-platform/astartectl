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

package pairing

import (
	"errors"
	"fmt"
	"os"

	"github.com/astarte-platform/astarte-go/misc"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage device registration",
}

var agentRegisterCmd = &cobra.Command{
	Use:   "register <device_id>",
	Short: "Register a device",
	Long: `Register a new device to your realm.

This returns the credentials_secret that can be use to obtain device credentials.
<device_id> must be a 128 bit base64 url-encoded UUID`,
	Example: `  astartectl pairing agent register 2TBn-jNESuuHamE2Zo1anA`,
	Args:    cobra.ExactArgs(1),
	RunE:    agentRegisterF,
}

var agentUnregisterCmd = &cobra.Command{
	Use:   "unregister <device_id>",
	Short: "Unregister a device",
	Long: `Unregister a device, making it possible to register it again even after it has requested its credentials.

All data belonging to the device will be kept as is in Astarte.`,
	Example: `  astartectl pairing agent unregister 2TBn-jNESuuHamE2Zo1anA`,
	Args:    cobra.ExactArgs(1),
	RunE:    agentUnregisterF,
}

func init() {
	agentUnregisterCmd.PersistentFlags().BoolP("non-interactive", "y", false, "Non-interactive mode. Will answer yes by default to all questions.")

	PairingCmd.AddCommand(agentCmd)

	agentCmd.AddCommand(
		agentRegisterCmd,
		agentUnregisterCmd,
	)
}

func agentRegisterF(command *cobra.Command, args []string) error {
	// TODO: add support for initial_introspection
	deviceID := args[0]
	if !misc.IsValidAstarteDeviceID(deviceID) {
		return errors.New("Invalid device id")
	}

	credentialsSecret, err := astarteAPIClient.Pairing.RegisterDevice(realm, deviceID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Print the Credentials Secret
	fmt.Printf("Device %s successfully registered in Realm %s.\n", deviceID, realm)
	fmt.Printf("The Device's Credentials Secret is \"%s\".\n", credentialsSecret)
	fmt.Println()
	fmt.Println("Please don't share the Credentials Secret, and ensure it is transferred securely to your Device.")
	fmt.Printf("Once the Device pairs for the first time, the Credentials Secret ")
	fmt.Printf("will be associated permanently to the Device and it won't be changeable anymore.\n")
	return nil
}

func agentUnregisterF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	if !misc.IsValidAstarteDeviceID(deviceID) {
		return errors.New("Invalid device id")
	}

	nonInteractive, err := command.Flags().GetBool("non-interactive")
	if err != nil {
		return err
	}

	fmt.Printf("Will unregister device %s from realm %s.\n", deviceID, realm)
	if !nonInteractive {
		confirmation, err := utils.AskForConfirmation("Do you want to continue?")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if !confirmation {
			return nil
		}
	}

	err = astarteAPIClient.Pairing.UnregisterDevice(realm, deviceID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}
