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
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"time"
)

// realmsCmd represents the realms command
var realmsCmd = &cobra.Command{
	Use:   "realms",
	Short: "Manage realms",
	Long:  `List, show or create realms in your Astarte instance.`,
}

var realmsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List realms",
	Long:    "List realms present in your Astarte instance.",
	Example: `  astartectl housekeeping realms list`,
	RunE:    realmsListF,
}

var realmsShowCmd = &cobra.Command{
	Use:     "show <realm_name>",
	Short:   "Show realm",
	Long:    "Show a realm in your Astarte instance.",
	Example: `  astartectl housekeeping realms show myrealm`,
	Args:    cobra.ExactArgs(1),
	RunE:    realmsShowF,
}

var realmsCreateCmd = &cobra.Command{
	Use:     "create <realm_name>",
	Short:   "Create realm",
	Long:    "Create a realm in your Astarte instance.",
	Example: `  astartectl housekeeping realms create myrealm -p /path/to/public_key`,
	Args:    cobra.ExactArgs(1),
	RunE:    realmsCreateF,
}

var netClient = &http.Client{
	Timeout: time.Second * 30,
}

func init() {
	HousekeepingCmd.AddCommand(realmsCmd)

	realmsCreateCmd.Flags().StringP("public-key", "p", "", "Path to PEM encoded public key used as realm key")
	realmsCreateCmd.MarkFlagRequired("public-key")
	realmsCreateCmd.MarkFlagFilename("public-key")
	realmsCreateCmd.Flags().IntP("replication-factor", "r", 1, "Replication factor for the realm, used with SimpleStrategy replication")

	realmsCmd.AddCommand(
		realmsListCmd,
		realmsShowCmd,
		realmsCreateCmd,
	)
}

func realmsListF(command *cobra.Command, args []string) error {
	req, err := http.NewRequest("GET", housekeepingUrl+"/v1/realms", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+housekeepingJwt)

	resp, err := netClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		var responseBody struct {
			Data []string `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			return err
		}

		respJson, _ := json.MarshalIndent(responseBody, "", "  ")
		fmt.Println(string(respJson))
	} else {
		var errorBody struct {
			Errors map[string]interface{} `json:"errors"`
		}
		err = json.NewDecoder(resp.Body).Decode(&errorBody)
		if err != nil {
			return err
		}

		errJson, _ := json.MarshalIndent(&errorBody, "", "  ")
		fmt.Println(string(errJson))
		return nil
	}

	return nil
}

func realmsShowF(command *cobra.Command, args []string) error {
	realm := args[0]

	fmt.Printf("Show realm called\nrealm: %s\n", realm)
	fmt.Printf("Going to call %s with this JWT %s\n", housekeepingUrl, housekeepingJwt)

	return nil
}

func realmsCreateF(command *cobra.Command, args []string) error {
	realm := args[0]
	publicKey, err := command.Flags().GetString("public-key")
	if err != nil {
		return err
	}

	replicationFactor, err := command.Flags().GetInt("replication-factor")
	if err != nil {
		return err
	}

	fmt.Printf("Create realms called\nrealm: %s, public_key: %s, replication factor: %d\n", realm, publicKey, replicationFactor)
	fmt.Printf("Going to call %s with this JWT %s\n", housekeepingUrl, housekeepingJwt)

	return nil
}
