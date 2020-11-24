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

// TODO: link to the doc when it's updated
var installDeprecationMessage = `Deprecated command.

Since Astarte 1.0, operator installation through astartectl has been removed. You can install astarte-operator using Helm (https://helm.sh/) with these commands:

$ helm repo add astarte https://helm.astarte-platform.org
$ helm repo update
$ helm install astarte-operator astarte/astarte-operator --version 1.0.0-alpha.1`

var installCmd = &cobra.Command{
	Use:   "install-operator",
	Short: "deprecated - See astartectl cluster install-operator -h",
	Long:  installDeprecationMessage,
	RunE:  clusterOperatorInstallF,
	// Ignore flags so we always print deprecation message
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
}

func init() {
	ClusterCmd.AddCommand(installCmd)
}

func clusterOperatorInstallF(command *cobra.Command, args []string) error {
	// Print deprecation message and exit with 1 so that scripts detect the failure
	fmt.Fprintln(os.Stderr, installDeprecationMessage)
	os.Exit(1)

	return nil
}
