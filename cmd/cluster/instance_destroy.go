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
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy <name>",
	Short: "Destroy an Astarte Instance in the current Kubernetes Cluster",
	Long: `Destroy an Astarte Instance in the current Kubernetes Cluster. This will adhere to the same current-context
kubectl mentions. Please be aware of the fact that when an Astarte instance is destroyed, there is no way to recover it.`,
	Example: `  astartectl cluster instances destroy astarte`,
	RunE:    clusterDestroyF,
	Args:    cobra.ExactArgs(1),
}

func init() {
	destroyCmd.PersistentFlags().Bool("delete-volumes", false, "When set, all the Persistent Volume Claims will be destroyed. All data will be lost with no means of recovery.")
	destroyCmd.PersistentFlags().BoolP("non-interactive", "y", false, "Non-interactive mode. Will answer yes by default to all questions.")

	InstancesCmd.AddCommand(destroyCmd)
}

func clusterDestroyF(command *cobra.Command, args []string) error {
	resourceName := args[0]
	resourceNamespace, err := command.Flags().GetString("namespace")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if resourceNamespace == "" {
		resourceNamespace = "astarte"
	}
	deleteVolumes, err := command.Flags().GetBool("delete-volumes")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if _, err := getAstarteInstance(resourceName, resourceNamespace); err != nil {
		fmt.Fprintf(os.Stderr, "Could not find resource %s in namespace %s.\n", resourceName, resourceNamespace)
		os.Exit(1)
	}
	y, err := command.Flags().GetBool("non-interactive")
	if err != nil {
		return err
	}

	fmt.Printf("Will destroy Astarte instance %s in namespace %s.\n", resourceName, resourceNamespace)
	fmt.Println("WARNING: This operation is NOT REVERSIBLE and ALL DATA WILL BE LOST!!!")
	confirmation, _ := utils.PromptChoice("To continue, please enter the exact name of the Astarte instance you are deleting:", "", true, y)
	if confirmation != resourceName {
		fmt.Fprintln(os.Stderr, "Aborting.")
		os.Exit(1)
	}

	fmt.Println("Destroying Astarte instance...")
	// Kill it.
	err = kubernetesDynamicClient.Resource(astarteV1Alpha1).Namespace(resourceNamespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while destroying Astarte Resource.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Your Astarte instance has been successfully destroyed. Please allow a few minutes for the Cluster to settle and for all resource to be evicted.")

	if deleteVolumes {
		fmt.Println("Deleting all Volumes...")
		pvcList, err := kubernetesClient.CoreV1().PersistentVolumeClaims(resourceNamespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not list PVCs. You might need to delete Volumes manually.")
			os.Exit(1)
		}

		for _, pvc := range pvcList.Items {
			// Check for all services which spawn a PVC, and delete it in case.
			for _, svc := range []string{"cfssl", "cassandra", "rabbitmq", "vernemq"} {
				if strings.HasPrefix(pvc.Name, resourceName+"-"+svc+"-data") {
					err = kubernetesClient.CoreV1().PersistentVolumeClaims(resourceNamespace).Delete(context.TODO(), pvc.Name, metav1.DeleteOptions{})
					if err != nil {
						fmt.Printf("WARNING: Persistent Volume Claim %s could not be deleted.\n", pvc.Name)
					}
					break
				}
			}
		}

		fmt.Println("Volumes Deleted.")
	}

	return nil
}
