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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	realmsCreateCmd.Flags().IntP("replication-factor", "r", 0, `Replication factor for the realm, used with SimpleStrategy replication.`)
	realmsCreateCmd.Flags().StringSliceP("datacenter-replication", "d", nil,
		`Replication factor for a datacenter, used with NetworkTopologyStrategy replication.

The format is <datacenter-name>:<replication-factor>,<other-datacenter-name>:<other-replication-factor>.
You can also specify the flag multiple times instead of separating it with a comma.`)

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
		os.Exit(1)
	}

	return nil
}

func realmsShowF(command *cobra.Command, args []string) error {
	realm := args[0]

	req, err := http.NewRequest("GET", housekeepingUrl+"/v1/realms/"+realm, nil)
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
			Data map[string]interface{} `json:"data"`
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
		os.Exit(1)
	}

	return nil
}

func realmsCreateF(command *cobra.Command, args []string) error {
	realm := args[0]
	publicKey, err := command.Flags().GetString("public-key")
	if err != nil {
		return err
	}

	publicKeyContent, err := ioutil.ReadFile(publicKey)
	if err != nil {
		return err
	}

	var requestBody struct {
		Data map[string]interface{} `json:"data"`
	}

	requestBody.Data = map[string]interface{}{
		"realm_name":         realm,
		"jwt_public_key_pem": string(publicKeyContent),
	}

	replicationFactor, err := command.Flags().GetInt("replication-factor")
	if err != nil {
		return err
	}

	datacenterReplications, err := command.Flags().GetStringSlice("datacenter-replication")
	if err != nil {
		return err
	}

	if replicationFactor > 0 && len(datacenterReplications) > 0 {
		return errors.New("replication-factor and datacenter-replication are mutually exclusive, you only have to specify one")
	}

	if replicationFactor > 0 {
		requestBody.Data["replication_factor"] = replicationFactor
	} else if len(datacenterReplications) > 0 {
		requestBody.Data["replication_class"] = "NetworkTopologyStrategy"
		datacenterReplicationFactors := make(map[string]int)
		for _, datacenterString := range datacenterReplications {
			tokens := strings.Split(datacenterString, ":")
			if len(tokens) != 2 {
				errString := "Invalid datacenter replication: " + datacenterString + "."
				errString += "\nFormat must be <datacenter-name>:<replication-factor>"
				return errors.New(errString)
			}
			datacenter := tokens[0]
			datacenterReplicationFactor, err := strconv.Atoi(tokens[1])
			if err != nil {
				return errors.New("Invalid replication factor " + tokens[1])
			}
			datacenterReplicationFactors[datacenter] = datacenterReplicationFactor
		}
		requestBody.Data["datacenter_replication_factors"] = datacenterReplicationFactors
	}

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", housekeepingUrl+"/v1/realms", b)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+housekeepingJwt)
	req.Header.Add("Content-Type", "application/json")

	resp, err := netClient.Do(req)
	if err != nil {
		return err
	}

	var responseBody struct {
		Errors map[string]interface{} `json:"errors"`
	}

	if resp.StatusCode == 201 {
		fmt.Println("ok")
	} else {
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			return err
		}

		errJson, _ := json.MarshalIndent(&responseBody, "", "  ")
		fmt.Println(string(errJson))
		os.Exit(1)
	}

	return nil
}
