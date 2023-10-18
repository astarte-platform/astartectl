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

	"github.com/astarte-platform/astarte-go/triggers"

	"github.com/spf13/cobra"
)

var triggersCmd = &cobra.Command{
	Use:   "triggers",
	Short: "Utility operations on Astarte Triggers",
}

var validateTriggerCmd = &cobra.Command{
	Use:   "validate <trigger_file>",
	Short: "Validates a trigger",
	Long: `Checks whether the provided JSON file is a valid Astarte Trigger.
Note that the checks performed by this function are not as thorough as the ones performed by Astarte, so there could be false positives (but no false negatives).
This command is thought to be used in CI pipelines to validate that new triggers are "reasonable enough".

Returns 0 and does not print anything if the trigger is valid, returns 1 and prints an error message if it isn't.`,
	Example: `  astartectl utils triggers validate my_trigger.json`,
	Args:    cobra.ExactArgs(1),
	RunE:    validateTriggerF,
}

func init() {
	UtilsCmd.AddCommand(triggersCmd)

	triggersCmd.AddCommand(
		validateTriggerCmd,
	)
}

func validateTriggerF(command *cobra.Command, args []string) error {
	triggerPath := args[0]

	if _, err := triggers.ParseTriggerFrom(triggerPath); err != nil {
		fmt.Fprintf(os.Stderr, "%s is not a valid Astarte Trigger: %s\n", triggerPath, err)
		os.Exit(1)
	}

	return nil
}
