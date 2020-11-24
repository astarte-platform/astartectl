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
var uninstallDeprecationMessage = `Deprecated command.

Since Astarte 1.0, operator uninstallation through astartectl has been removed. You can uninstall astarte-operator using Helm (https://helm.sh/) with this command:

$ helm uninstall astarte-operator

The above command will not remove Astarte CRDs, if you want to remove them (keeping in mind that removing them is DANGEROUS and can lead to data loss in your cluster) you can do so with these commands:

$ kubectl delete crd astartes.api.astarte-platform.org
$ kubectl delete crd astartevoyageringresses.api.astarte-platform.org`

var uninstallCmd = &cobra.Command{
	Use:   "uninstall-operator",
	Short: "deprecated - See astartectl cluster uninstall-operator -h",
	Long:  uninstallDeprecationMessage,
	RunE:  clusterOperatorUninstallF,
	// Ignore flags so we always print deprecation message
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
}

func init() {
	ClusterCmd.AddCommand(uninstallCmd)
}

func clusterOperatorUninstallF(command *cobra.Command, args []string) error {
	// Print deprecation message and exit with 1 so that scripts detect the failure
	fmt.Fprintln(os.Stderr, uninstallDeprecationMessage)
	os.Exit(1)

	return nil
}
