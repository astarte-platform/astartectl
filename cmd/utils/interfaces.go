// Copyright © 2022 SECO Mind Srl
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
	"fmt"
	"os"

	"github.com/astarte-platform/astarte-go/interfaces"

	"github.com/spf13/cobra"
)

var interfacesCmd = &cobra.Command{
	Use:   "interfaces",
	Short: "Utility operations on Astarte Interfaces",
}

var validateInterfaceCmd = &cobra.Command{
	Use:   "validate <interface_file>",
	Short: "Validates an interface",
	Long: `Checks whether the provided JSON file is a valid Astarte Interface.
Note that the checks performed by this function are not as thorough as the ones performed by Astarte, so there could be false positives (but no false negatives).
This command is thought to be used in CI pipelines to validate that new interfaces are "reasonable enough".

Returns 0 and does not print anything if the interface is valid, returns 1 and prints an error message if it isn't.`,
	Example: `  astartectl utils interfaces validate com.my.Interface.json`,
	Args:    cobra.ExactArgs(1),
	RunE:    validateInterfaceF,
}

func init() {
	UtilsCmd.AddCommand(interfacesCmd)

	interfacesCmd.AddCommand(
		validateInterfaceCmd,
	)
}

func validateInterfaceF(command *cobra.Command, args []string) error {
	interfacePath := args[0]

	if _, err := interfaces.ParseInterfaceFromFile(interfacePath); err != nil {
		fmt.Fprintf(os.Stderr, "%s is not a valid Astarte Interface: %s\n", interfacePath, err)
		os.Exit(1)
	}

	return nil
}
