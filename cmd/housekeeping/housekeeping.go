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
	"github.com/astarte-platform/astarte-go/client"
	"github.com/astarte-platform/astarte-go/misc"
	"github.com/astarte-platform/astartectl/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// HousekeepingCmd represents the housekeeping command
var HousekeepingCmd = &cobra.Command{
	Use:               "housekeeping",
	Short:             "Interact with Housekeeping API",
	Long:              `Interact with Housekeeping API.`,
	PersistentPreRunE: housekeepingPersistentPreRunE,
}

var astarteAPIClient *client.Client

func init() {
	HousekeepingCmd.PersistentFlags().StringP("housekeeping-key", "k", "",
		"Path to housekeeping private key to generate JWT for authentication")
	HousekeepingCmd.MarkPersistentFlagFilename("housekeeping-key")
	viper.BindPFlag("housekeeping.key-file", HousekeepingCmd.PersistentFlags().Lookup("housekeeping-key"))
	HousekeepingCmd.PersistentFlags().String("housekeeping-url", "",
		"Housekeeping API base URL. Defaults to <astarte-url>/housekeeping.")
	viper.BindPFlag("individual-urls.housekeeping", HousekeepingCmd.PersistentFlags().Lookup("housekeeping-url"))
}

func housekeepingPersistentPreRunE(cmd *cobra.Command, args []string) error {
	var err error
	astarteAPIClient, err = utils.APICommandSetup(map[misc.AstarteService]string{misc.Housekeeping: "individual-urls.housekeeping"},
		"housekeeping.key", "housekeeping.key-file")

	return err
}
