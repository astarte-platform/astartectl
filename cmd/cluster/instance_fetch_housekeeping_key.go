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

package cluster

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var fetchHKPrivateKeyCmd = &cobra.Command{
	Use:     "fetch-housekeeping-key <name>",
	Short:   "Fetches Housekeeping private key from the specified instance",
	Long:    `Fetches Housekeeping private key from the specified instance.`,
	Example: `  astartectl cluster fetch-housekeeping-private-key`,
	RunE:    fetchHKPrivateKeyF,
	Args:    cobra.ExactArgs(1),
}

func init() {
	fetchHKPrivateKeyCmd.PersistentFlags().StringP("output", "o", "", "When specified, saves the key to the specified file rather than printing it in stdout.")

	InstancesCmd.AddCommand(fetchHKPrivateKeyCmd)
}

func fetchHKPrivateKeyF(command *cobra.Command, args []string) error {
	resourceName := args[0]
	resourceNamespace, err := command.Flags().GetString("namespace")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if resourceNamespace == "" {
		resourceNamespace = "astarte"
	}

	keyData, err := getHousekeepingKey(resourceName, resourceNamespace, true)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	outputFile, err := command.Flags().GetString("output")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if outputFile == "" {
		fmt.Print(string(keyData))
	} else {
		outFile, err := os.Create(outputFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer outFile.Close()

		if _, err := outFile.Write(keyData); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	return nil
}
