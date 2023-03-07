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
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var devicesCredentialsCmd = &cobra.Command{
	Use:   "credentials",
	Short: "Manage device credentials status",
}

var devicesCredentialsInhibitCmd = &cobra.Command{
	Use:   "inhibit <device_id_or_alias> <true|false>",
	Short: "Enable or disable credentials inhibition for a device",
	Long: `Enable or disable credentials inhibition for a device.
An inhibited device can't request new credentials to Pairing API
If you pass true as second parameter the device will be inhibited, if you pass false it will be able to request
credentials again. Note that inhibiting a device does not revoke its current credentials, they will remain valid
until their expiration.`,
	Example: `  astartectl appengine devices credentials inhibit 2TBn-jNESuuHamE2Zo1anA true`,
	Args:    cobra.ExactArgs(2),
	RunE:    devicesCredentialsInhibitF,
}

func init() {
	devicesCmd.AddCommand(
		devicesCredentialsCmd,
	)

	devicesCredentialsCmd.AddCommand(
		devicesCredentialsInhibitCmd,
	)

	devicesCredentialsCmd.PersistentFlags().String("force-id-type", "",
		"When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")
}

func devicesCredentialsInhibitF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	forceIDType, err := command.Flags().GetString("force-id-type")
	if err != nil {
		return err
	}
	deviceIdentifierType, err := deviceIdentifierTypeFromFlags(deviceID, forceIDType)
	if err != nil {
		return err
	}

	inhibitString := args[1]
	inhibit, err := strconv.ParseBool(inhibitString)
	if err != nil {
		return errors.New("The second argument should be one of: [true false]")
	}

	inhibitDeviceReq, err := astarteAPIClient.SetDeviceInhibited(realm, deviceID, deviceIdentifierType, inhibit)
	if err != nil {
		return err
	}
	inhibitDeviceRes, err := inhibitDeviceReq.Run(astarteAPIClient)
	if err != nil {
		return err
	}
	_, _ = inhibitDeviceRes.Parse()

	fmt.Println("ok")
	return nil
}
