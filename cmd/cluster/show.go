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
	"strings"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:     "show",
	Short:   "Shows existing Astarte Cluster and Kubernetes Context",
	Long:    `Shows the existing Astarte Cluster, Operator and Kubernetes state.`,
	Example: `  astartectl cluster show`,
	RunE:    clusterShowF,
}

func init() {
	ClusterCmd.AddCommand(showCmd)
}

func clusterShowF(command *cobra.Command, args []string) error {
	operator, err := getAstarteOperator()
	if err != nil {
		fmt.Println("Could not find an Astarte Operator Deployment on this Kubernetes Cluster.")
		fmt.Println()
		fmt.Println("To install Astarte Operator in this cluster, please run astartectl cluster install-operator.")
		os.Exit(0)
	}

	fmt.Printf("This Cluster is running Astarte Operator version %s.\n\n",
		strings.Split(operator.Spec.Template.Spec.Containers[0].Image, ":")[1])

	astartes, err := listAstartes()
	if err != nil || len(astartes) == 0 {
		fmt.Println("No Managed Astarte installations found. Maybe you want to deploy one with astartectl cluster instance deploy?")
		return nil
	}

	fmt.Println("Managed Astarte Instances:")
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"Name", "Namespace", "Version", "Deployment Profile", "Operator Status"})

	for _, v := range astartes {
		for _, res := range v.Items {
			operatorStatus, _, deploymentProfile := getManagedAstarteResourceStatus(res)

			t.AppendRow(table.Row{res.Object["metadata"].(map[string]interface{})["name"],
				res.Object["metadata"].(map[string]interface{})["namespace"],
				res.Object["spec"].(map[string]interface{})["version"],
				deploymentProfile, operatorStatus,
			})
		}
	}

	t.Render()

	return nil
}
