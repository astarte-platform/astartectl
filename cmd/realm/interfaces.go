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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// interfacesCmd represents the interfaces command
var interfacesCmd = &cobra.Command{
	Use:   "interfaces",
	Short: "Manage interfaces",
	Long:  `List, show, install or update interfaces in your realm.`,
}

var interfacesListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List interfaces",
	Long:    `List the name of the interfaces installed in the realm.`,
	Example: `  astartectl realm-management interfaces list`,
	RunE:    interfacesListF,
}

var interfacesVersionsCmd = &cobra.Command{
	Use:     "versions <interface_name>",
	Short:   "List major versions of an interface",
	Long:    `List the major versions of an interface installed in the realm.`,
	Example: `  astartectl realm-management interfaces versions com.my.Interface`,
	Args:    cobra.ExactArgs(1),
	RunE:    interfacesVersionsF,
}

var interfacesShowCmd = &cobra.Command{
	Use:     "show <interface_name> <interface_major>",
	Short:   "Show interface",
	Long:    `Show the given major version of the interface installed in the realm.`,
	Example: `  astartectl realm-management interfaces show com.my.Interface 0`,
	Args:    cobra.ExactArgs(2),
	RunE:    interfacesShowF,
}

var interfacesInstallCmd = &cobra.Command{
	Use:   "install <interface_file>",
	Short: "Install interface",
	Long: `Install the given interface in the realm.
<interface_file> must be a path to a JSON file containing a valid Astarte interface.`,
	Example: `  astartectl realm-management interfaces install com.my.Interface.json`,
	Args:    cobra.ExactArgs(1),
	RunE:    interfacesInstallF,
}

var interfacesUpdateCmd = &cobra.Command{
	Use:   "update <interface_file>",
	Short: "Update interface",
	Long: `Update the given interface in the realm.
<interface_file> must be a path to a JSON file containing a valid Astarte interface.

The name and major version of the interface are read from the interface file.`,
	Example: `  astartectl realm-management interfaces update com.my.Interface.json`,
	Args:    cobra.ExactArgs(1),
	RunE:    interfacesUpdateF,
}

var netClient = &http.Client{
	Timeout: time.Second * 30,
}

func init() {
	RealmManagementCmd.AddCommand(interfacesCmd)

	interfacesCmd.AddCommand(
		interfacesListCmd,
		interfacesVersionsCmd,
		interfacesShowCmd,
		interfacesInstallCmd,
		interfacesUpdateCmd,
	)
}

func interfacesListF(command *cobra.Command, args []string) error {
	realmInterfaces, err := astarteAPIClient.RealmManagement.ListInterfaces(realm, realmManagementJwt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return nil
	}

	fmt.Println(realmInterfaces)
	return nil
}

func interfacesVersionsF(command *cobra.Command, args []string) error {
	interfaceName := args[0]

	req, err := http.NewRequest("GET", realmManagementUrl+"/v1/"+realm+"/interfaces/"+interfaceName, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+realmManagementJwt)

	resp, err := netClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		var responseBody struct {
			Data []int `json:"data"`
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

func interfacesShowF(command *cobra.Command, args []string) error {
	interfaceName := args[0]
	interfaceMajor := args[1]

	req, err := http.NewRequest("GET", realmManagementUrl+"/v1/"+realm+"/interfaces/"+interfaceName+"/"+interfaceMajor, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+realmManagementJwt)

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

func interfacesInstallF(command *cobra.Command, args []string) error {
	interfaceFile, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	var interfaceBody map[string]interface{}
	err = json.Unmarshal(interfaceFile, &interfaceBody)
	if err != nil {
		return err
	}

	var requestBody struct {
		Data map[string]interface{} `json:"data"`
	}
	requestBody.Data = interfaceBody

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", realmManagementUrl+"/v1/"+realm+"/interfaces", b)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+realmManagementJwt)
	req.Header.Add("Content-Type", "application/json")

	resp, err := netClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 201 {
		fmt.Println("ok")
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

func interfacesUpdateF(command *cobra.Command, args []string) error {
	interfaceFile, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	var interfaceBody map[string]interface{}
	err = json.Unmarshal(interfaceFile, &interfaceBody)
	if err != nil {
		return err
	}

	interfaceName := fmt.Sprintf("%v", interfaceBody["interface_name"])
	interfaceMajor := fmt.Sprintf("%v", interfaceBody["version_major"])

	var requestBody struct {
		Data map[string]interface{} `json:"data"`
	}
	requestBody.Data = interfaceBody

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", realmManagementUrl+"/v1/"+realm+"/interfaces/"+interfaceName+"/"+interfaceMajor, b)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+realmManagementJwt)
	req.Header.Add("Content-Type", "application/json")

	resp, err := netClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 201 {
		fmt.Println("ok")
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
