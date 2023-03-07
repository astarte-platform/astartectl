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

	"github.com/astarte-platform/astarte-go/astarteservices"
	"github.com/astarte-platform/astarte-go/client"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// PairingCmd represents the pairing command
var PairingCmd = &cobra.Command{
	Use:               "pairing",
	Short:             "Interact with Pairing API",
	Long:              `Interact with pairing API to register devices or to work with device credentials`,
	PersistentPreRunE: pairingPersistentPreRunE,
}

var realm string
var astarteAPIClient *client.Client

func init() {
	PairingCmd.PersistentFlags().StringP("realm-key", "k", "",
		"Path to realm private key used to generate JWT for authentication")
	PairingCmd.MarkPersistentFlagFilename("realm-key")
	PairingCmd.PersistentFlags().String("pairing-url", "",
		"Pairing API base URL. Defaults to <astarte-url>/pairing.")
	viper.BindPFlag("individual-urls.pairing", PairingCmd.PersistentFlags().Lookup("pairing-url"))
	PairingCmd.PersistentFlags().StringP("realm-name", "r", "",
		"The name of the realm that will be queried")
	PairingCmd.PersistentFlags().Bool("to-curl", false,
		"When set, display a command-line equivalent instead of running the command.")
	viper.BindPFlag("pairing-to-curl", PairingCmd.PersistentFlags().Lookup("to-curl"))
}

func pairingPersistentPreRunE(cmd *cobra.Command, args []string) error {
	viper.BindPFlag("realm.key-file", cmd.Flags().Lookup("realm-key"))
	var err error
	astarteAPIClient, err = utils.APICommandSetup(
		map[astarteservices.AstarteService]string{astarteservices.Pairing: "individual-urls.pairing"}, "realm.key", "realm.key-file")
	if err != nil {
		return err
	}

	viper.BindPFlag("realm.name", cmd.Flags().Lookup("realm-name"))
	realm = viper.GetString("realm.name")
	if realm == "" {
		return errors.New("realm is required")
	}

	// if just --to-curl is given, default to true
	cmd.Flags().Lookup("to-curl").NoOptDefVal = "true"

	return nil
}
