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
	"github.com/astarte-platform/astarte-go/deviceid"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
)

// AliasesCmd represents the aliases command
var aliasesCmd = &cobra.Command{
	Use:     "aliases",
	Short:   "Interact with Device Aliases",
	Long:    `Perform actions on Astarte Device Aliases.`,
	Aliases: []string{"alias"},
}

var aliasesListCmd = &cobra.Command{
	Use:     "list <device_id>",
	Short:   "List aliases",
	Long:    `List all aliases for a device.`,
	Example: `  astartectl appengine devices aliases list`,
	Args:    cobra.ExactArgs(1),
	RunE:    aliasesListF,
	Aliases: []string{"ls"},
}

var aliasesAddCmd = &cobra.Command{
	Use:     "add <device_id> <alias-tag>=<alias>",
	Short:   "Add an Alias",
	Long:    `Adds an Alias to the Device with ID <device_id>, in the form <alias-tag>=<alias>.`,
	Example: `  astartectl appengine devices aliases add 2TBn-jNESuuHamE2Zo1anA my-alias-tag=device12345`,
	Args:    cobra.ExactArgs(2),
	RunE:    aliasesAddF,
}

var aliasesRemoveCmd = &cobra.Command{
	Use:     "remove <device_id> <alias_tag>",
	Short:   "Remove an Alias from a Device",
	Long:    `Removes an Alias from the Device with ID <device_id>, by specifying its tag.`,
	Example: `  astartectl appengine devices aliases remove 2TBn-jNESuuHamE2Zo1anA my-alias-tag`,
	Args:    cobra.ExactArgs(2),
	RunE:    aliasesRemoveF,
	Aliases: []string{"rm"},
}

func init() {
	devicesCmd.AddCommand(aliasesCmd)

	aliasesCmd.AddCommand(
		aliasesListCmd,
		aliasesAddCmd,
		aliasesRemoveCmd,
	)
}

func aliasesListF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	if !deviceid.IsValid(deviceID) {
		fmt.Fprintf(os.Stderr, "%s is not a valid Astarte Device ID\n", deviceID)
		os.Exit(1)
	}
	deviceAliasesCall, err := astarteAPIClient.ListDeviceAliases(realm, deviceID, client.AstarteDeviceID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	utils.MaybeCurlAndExit(deviceAliasesCall, astarteAPIClient)

	deviceAliasesRes, err := deviceAliasesCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	aliases, _ := deviceAliasesRes.Parse()
	aliasesMap, _ := aliases.(map[string]string)

	for k, v := range aliasesMap {
		fmt.Printf("%v: %v\n", k, v)
	}

	return nil
}

func aliasesAddF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	if !deviceid.IsValid(deviceID) {
		fmt.Fprintf(os.Stderr, "%s is not a valid Astarte Device ID\n", deviceID)
		os.Exit(1)
	}
	alias := args[1]
	s := strings.Split(alias, "=")
	if len(s) != 2 {
		fmt.Fprintf(os.Stderr, "Alias should be in the form <alias-tag>=<alias>")
		os.Exit(1)
	}

	addDeviceCall, err := astarteAPIClient.AddDeviceAlias(realm, deviceID, s[0], s[1])
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

func aliasesRemoveF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	if !deviceid.IsValid(deviceID) {
		fmt.Fprintf(os.Stderr, "%s is not a valid Astarte Device ID\n", deviceID)
		os.Exit(1)
	}
	aliasTag := args[1]

	deleteAliasCall, err := astarteAPIClient.DeleteDeviceAlias(realm, deviceID, aliasTag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	utils.MaybeCurlAndExit(deleteAliasCall, astarteAPIClient)

	deleteAliasRes, err := deleteAliasCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, _ = deleteAliasRes.Parse()

	fmt.Println("ok")
	return nil
}
