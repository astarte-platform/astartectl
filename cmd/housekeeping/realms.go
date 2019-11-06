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
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// realmsCmd represents the realms command
var realmsCmd = &cobra.Command{
	Use:     "realms",
	Short:   "Manage realms",
	Long:    `List, show or create realms in your Astarte instance.`,
	Aliases: []string{"realm"},
}

var realmsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List realms",
	Long:    "List realms present in your Astarte instance.",
	Example: `  astartectl housekeeping realms list`,
	RunE:    realmsListF,
	Aliases: []string{"ls"},
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
	realms, err := astarteAPIClient.Housekeeping.ListRealms(housekeepingJwt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(realms)
	return nil
}

func realmsShowF(command *cobra.Command, args []string) error {
	realm := args[0]

	realmDetails, err := astarteAPIClient.Housekeeping.GetRealm(realm, housekeepingJwt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("%+v\n", realmDetails)
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
		err = astarteAPIClient.Housekeeping.CreateRealmWithReplicationFactor(realm, string(publicKeyContent), replicationFactor, housekeepingJwt)
	} else if len(datacenterReplications) > 0 {
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
		err = astarteAPIClient.Housekeeping.CreateRealmWithDatacenterReplication(realm, string(publicKeyContent),
			datacenterReplicationFactors, housekeepingJwt)
	} else {
		err = astarteAPIClient.Housekeeping.CreateRealm(realm, string(publicKeyContent), housekeepingJwt)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("ok")
	return nil
}
