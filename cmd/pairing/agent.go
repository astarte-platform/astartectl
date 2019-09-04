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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"time"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage device registration",
}

var agentRegisterCmd = &cobra.Command{
	Use:   "register <device_id>",
	Short: "Register a device",
	Long: `Register a new device to your realm.

This returns the credentials_secret that can be use to obtain device credentials.
<device_id> must be a 128 bit base64 url-encoded UUID`,
	Example: `  astartectl pairing agent register 2TBn-jNESuuHamE2Zo1anA`,
	Args:    cobra.ExactArgs(1),
	RunE:    agentRegisterF,
}

var netClient = &http.Client{
	Timeout: time.Second * 30,
}

func init() {
	PairingCmd.AddCommand(agentCmd)

	agentCmd.AddCommand(
		agentRegisterCmd,
	)
}

func isValidDeviceId(deviceId string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(deviceId)
	if err != nil {
		return false
	}

	// 16 bytes == 128 bit
	if len(decoded) != 16 {
		return false
	}

	return true
}

func agentRegisterF(command *cobra.Command, args []string) error {
	deviceId := args[0]
	if !isValidDeviceId(deviceId) {
		return errors.New("Invalid device id")
	}

	// TODO: add support for initial_introspection
	var requestBody struct {
		Data struct {
			HwId string `json:"hw_id"`
		} `json:"data"`
	}
	requestBody.Data.HwId = deviceId

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", pairingUrl+"/v1/"+realm+"/agent/devices", b)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+pairingJwt)
	req.Header.Add("Content-Type", "application/json")

	resp, err := netClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 201 {
		var responseBody struct {
			Data map[string]interface{} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			return err
		}

		respJson, _ := json.MarshalIndent(&responseBody, "", "  ")
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
