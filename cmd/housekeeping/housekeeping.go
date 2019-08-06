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

package housekeeping

import (
	"errors"
	"github.com/spf13/cobra"
	"path"
)

// housekeepingCmd represents the housekeeping command
var HousekeepingCmd = &cobra.Command{
	Use:               "housekeeping",
	Short:             "Interact with Housekeeping API",
	Long:              `Interact with Housekeeping API.`,
	PersistentPreRunE: housekeepingPersistentPreRunE,
}

var housekeepingKey string
var housekeepingUrl string

func init() {
	HousekeepingCmd.PersistentFlags().StringVarP(
		&housekeepingKey, "housekeeping-key", "k", "",
		"Path to housekeeping private key to generate JWT for authentication")
	HousekeepingCmd.MarkPersistentFlagFilename("housekeeping-key")
	HousekeepingCmd.PersistentFlags().String("housekeeping-url", "",
		"Housekeeping API base URL. Defaults to <astarte-url>/housekeeping.")
}

func housekeepingPersistentPreRunE(cmd *cobra.Command, args []string) error {
	housekeepingUrlFlag, errHousekeeping := cmd.Flags().GetString("housekeeping-url")
	astarteUrlFlag, errAstarte := cmd.Flags().GetString("astarte-url")

	if errHousekeeping == nil && housekeepingUrlFlag != "" {
		// Use explicit housekeeping-url
		housekeepingUrl = housekeepingUrlFlag
	} else if errAstarte == nil && astarteUrlFlag != "" {
		housekeepingUrl = path.Join(astarteUrlFlag, "housekeeping")
	} else {
		return errors.New("Either astarte-url or housekeeping-url have to be specified")
	}

	if housekeepingKey == "" {
		return errors.New("housekeeping-key is required")
	}

	return nil
}
