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

package realm

import (
	"errors"
	"net/url"
	"path"

	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// realmManagementCmd represents the realmManagement command
var RealmManagementCmd = &cobra.Command{
	Use:               "realm-management",
	Short:             "Interact with Realm Management API",
	Long:              `Interact with Realm Management API.`,
	PersistentPreRunE: realmManagementPersistentPreRunE,
}

var realm string
var realmManagementJwt string
var realmManagementUrl string

func init() {
	RealmManagementCmd.PersistentFlags().StringP("realm-key", "k", "",
		"Path to realm private key used to generate JWT for authentication")
	RealmManagementCmd.MarkPersistentFlagFilename("realm-key")
	RealmManagementCmd.PersistentFlags().String("realm-management-url", "",
		"Realm Management API base URL. Defaults to <astarte-url>/realmmanagement.")
	viper.BindPFlag("realm-management.url", RealmManagementCmd.PersistentFlags().Lookup("realm-management-url"))
	RealmManagementCmd.PersistentFlags().StringP("realm-name", "r", "",
		"The name of the realm that will be queried")
}

func realmManagementPersistentPreRunE(cmd *cobra.Command, args []string) error {
	realmManagementUrlOverride := viper.GetString("realm-management.url")
	astarteUrl := viper.GetString("url")
	if realmManagementUrlOverride != "" {
		// Use explicit realm-management-url
		realmManagementUrl = realmManagementUrlOverride
	} else if astarteUrl != "" {
		url, _ := url.Parse(astarteUrl)
		url.Path = path.Join(url.Path, "realmmanagement")
		realmManagementUrl = url.String()
	} else {
		return errors.New("Either astarte-url or realm-management-url have to be specified")
	}

	viper.BindPFlag("realm.key", cmd.Flags().Lookup("realm-key"))
	realmManagementKey := viper.GetString("realm.key")
	if realmManagementKey == "" {
		return errors.New("realm-key is required")
	}

	viper.BindPFlag("realm.name", cmd.Flags().Lookup("realm-name"))
	realm = viper.GetString("realm.name")
	if realm == "" {
		return errors.New("realm is required")
	}

	var err error
	realmManagementJwt, err = generateRealmManagementJWT(realmManagementKey)
	if err != nil {
		return err
	}

	return nil
}

func generateRealmManagementJWT(privateKey string) (jwtString string, err error) {
	return utils.GenerateAstarteJWTFromKeyFile(privateKey, utils.RealmManagement, nil, 300)
}
