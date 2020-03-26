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

	"github.com/astarte-platform/astarte-go/client"
	"github.com/astarte-platform/astarte-go/misc"
	"github.com/astarte-platform/astartectl/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AppEngineCmd represents the appEngine command
var AppEngineCmd = &cobra.Command{
	Use:               "appengine",
	Short:             "Interact with AppEngine API",
	Long:              `Interact with AppEngine API.`,
	PersistentPreRunE: appEnginePersistentPreRunE,
}

var realm string
var astarteAPIClient *client.Client

func init() {
	AppEngineCmd.PersistentFlags().StringP("realm-key", "k", "",
		"Path to realm private key used to generate JWT for authentication")
	AppEngineCmd.MarkPersistentFlagFilename("realm-key")
	AppEngineCmd.PersistentFlags().String("appengine-url", "",
		"AppEngine API base URL. Defaults to <astarte-url>/appengine.")
	viper.BindPFlag("appengine.url", AppEngineCmd.PersistentFlags().Lookup("appengine-url"))
	AppEngineCmd.PersistentFlags().String("realm-management-url", "",
		"Realm Management API base URL. Defaults to <astarte-url>/realmmanagement.")
	AppEngineCmd.PersistentFlags().StringP("realm-name", "r", "",
		"The name of the realm that will be queried")
}

func appEnginePersistentPreRunE(cmd *cobra.Command, args []string) error {
	appEngineURLOverride := viper.GetString("appengine.url")
	viper.BindPFlag("realm-management.url", cmd.Flags().Lookup("realm-management-url"))
	realmManagementURLOverride := viper.GetString("realm-management.url")
	// Handle a special failure case, if realm-management is provided but appengine isn't
	if appEngineURLOverride == "" && realmManagementURLOverride != "" {
		return errors.New("Either astarte-url or appengine-url have to be specified")
	}

	individualURLVariables := map[misc.AstarteService]string{
		misc.AppEngine:       "appengine.url",
		misc.RealmManagement: "realm-management.url",
	}

	viper.BindPFlag("realm.key", cmd.Flags().Lookup("realm-key"))
	var err error
	astarteAPIClient, err = utils.APICommandSetup(individualURLVariables, "realm.key")
	if err != nil {
		return err
	}

	viper.BindPFlag("realm.name", cmd.Flags().Lookup("realm-name"))
	realm = viper.GetString("realm.name")
	if realm == "" {
		return errors.New("realm is required")
	}

	return nil
}

func deviceIdentifierTypeFromFlags(deviceIdentifier string, forceDeviceIdentifier string) (client.DeviceIdentifierType, error) {
	switch forceDeviceIdentifier {
	case "":
		return client.AutodiscoverDeviceIdentifier, nil
	case "device-id":
		if !misc.IsValidAstarteDeviceID(deviceIdentifier) {
			return 0, fmt.Errorf("Required to evaluate the Device Identifier as an Astarte Device ID, but %v isn't a valid one", deviceIdentifier)
		}
		return client.AstarteDeviceID, nil
	case "alias":
		return client.AstarteDeviceAlias, nil
	}

	return 0, fmt.Errorf("%v is not a valid Astarte Device Identifier type. Valid options are [device-id alias]", forceDeviceIdentifier)
}
