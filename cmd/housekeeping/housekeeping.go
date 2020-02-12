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

	"github.com/astarte-platform/astartectl/client"

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

var housekeepingJwt string
var astarteAPIClient *client.Client

func init() {
	HousekeepingCmd.PersistentFlags().StringP("housekeeping-key", "k", "",
		"Path to housekeeping private key to generate JWT for authentication")
	HousekeepingCmd.MarkPersistentFlagFilename("housekeeping-key")
	viper.BindPFlag("housekeeping.key", HousekeepingCmd.PersistentFlags().Lookup("housekeeping-key"))
	HousekeepingCmd.PersistentFlags().String("housekeeping-url", "",
		"Housekeeping API base URL. Defaults to <astarte-url>/housekeeping.")
	viper.BindPFlag("housekeeping.url", HousekeepingCmd.PersistentFlags().Lookup("housekeeping-url"))
}

func housekeepingPersistentPreRunE(cmd *cobra.Command, args []string) error {
	housekeepingURLOverride := viper.GetString("housekeeping.url")
	astarteURL := viper.GetString("url")
	if housekeepingURLOverride != "" {
		// Use explicit housekeeping-url
		var err error
		astarteAPIClient, err = client.NewClientWithIndividualURLs("", housekeepingURLOverride, "", "", nil)
		if err != nil {
			return err
		}
	} else if astarteURL != "" {
		var err error
		astarteAPIClient, err = client.NewClient(astarteURL, nil)
		if err != nil {
			return err
		}
	} else {
		return errors.New("Either astarte-url or housekeeping-url have to be specified")
	}

	housekeepingKey := viper.GetString("housekeeping.key")
	explicitToken := viper.GetString("token")
	if housekeepingKey == "" && explicitToken == "" {
		return errors.New("housekeeping-key or token is required")
	}

	if explicitToken == "" {
		var err error
		housekeepingJwt, err = generateHousekeepingJWT(housekeepingKey)
		if err != nil {
			return err
		}
	} else {
		housekeepingJwt = explicitToken
	}

	return nil
}

func generateHousekeepingJWT(privateKey string) (jwtString string, err error) {
	return utils.GenerateAstarteJWTFromKeyFile(privateKey, map[utils.AstarteService][]string{utils.Housekeeping: []string{}}, 300)
}
