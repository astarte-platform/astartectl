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
	Use:     "list",
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
	if !utils.IsValidAstarteDeviceID(deviceID) {
		fmt.Printf("%s is not a valid Astarte Device ID\n", deviceID)
		os.Exit(1)
	}
	aliases, err := astarteAPIClient.AppEngine.ListDeviceAliases(realm, deviceID, appEngineJwt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("%v\n", aliases)
	return nil
}

func aliasesAddF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	if !utils.IsValidAstarteDeviceID(deviceID) {
		fmt.Printf("%s is not a valid Astarte Device ID\n", deviceID)
		os.Exit(1)
	}
	alias := args[1]
	s := strings.Split(alias, "=")
	if len(s) != 2 {
		fmt.Println("Alias should be in the form <alias-tag>=<alias>")
		os.Exit(1)
	}

	err := astarteAPIClient.AppEngine.AddDeviceAlias(realm, deviceID, s[0], s[1], appEngineJwt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}

func aliasesRemoveF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	if !utils.IsValidAstarteDeviceID(deviceID) {
		fmt.Printf("%s is not a valid Astarte Device ID\n", deviceID)
		os.Exit(1)
	}
	aliasTag := args[1]

	err := astarteAPIClient.AppEngine.DeleteDeviceAlias(realm, deviceID, aliasTag, appEngineJwt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}
