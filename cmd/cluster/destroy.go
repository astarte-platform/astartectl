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

	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy <name>",
	Short: "Destroy an Astarte Instance in the current Kubernetes Cluster",
	Long: `Destroy an Astarte Instance in the current Kubernetes Cluster. This will adhere to the same current-context
kubectl mentions. Please be aware of the fact that when an Astarte instance is destroyed, there is no way to recover it.`,
	Example: `  astartectl cluster destroy astarte`,
	RunE:    clusterDestroyF,
	Args:    cobra.ExactArgs(1),
}

func init() {
	destroyCmd.PersistentFlags().String("namespace", "", "Namespace in which to look for the Astarte resource will be destroyed.")
	destroyCmd.PersistentFlags().Bool("delete-volumes", false, "When set, all the Persistent Volume Claims will be destroyed. All data will be lost with no means of recovery.")
	destroyCmd.PersistentFlags().BoolP("non-interactive", "y", false, "Non-interactive mode. Will answer yes by default to all questions.")

	ClusterCmd.AddCommand(destroyCmd)
}

func clusterDestroyF(command *cobra.Command, args []string) error {
	astartes, err := listAstartes()
	if err != nil || len(astartes) == 0 {
		fmt.Println("No Managed Astarte installations found.")
		return nil
	}

	resourceName := args[0]
	resourceNamespace, err := command.Flags().GetString("namespace")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if resourceNamespace == "" {
		resourceNamespace = "astarte"
	}
	deleteVolumes, err := command.Flags().GetBool("delete-volumes")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var astarteObject *unstructured.Unstructured = nil
	for _, v := range astartes {
		for _, res := range v.Items {
			if res.Object["metadata"].(map[string]interface{})["namespace"] == resourceNamespace && res.Object["metadata"].(map[string]interface{})["name"] == resourceName {
				astarteObject = res.DeepCopy()
				break
			}
		}
	}

	if astarteObject == nil {
		fmt.Printf("Could not find resource %s in namespace %s.\n", resourceName, resourceNamespace)
		os.Exit(1)
	}

	fmt.Printf("Will destroy Astarte instance %s in namespace %s.\n", resourceName, resourceNamespace)
	fmt.Println("WARNING: This operation is NOT REVERSIBLE and ALL DATA WILL BE LOST!!!")
	confirmation, err := utils.PromptChoice("To continue, please enter the exact name of the Astarte instance you are deleting:", "", true)
	if confirmation != resourceName {
		fmt.Println("Aborting.")
		os.Exit(1)
	}

	fmt.Println("Destroying Astarte instance...")
	// Kill it.
	err = kubernetesDynamicClient.Resource(astarteV1Alpha1).Namespace(resourceNamespace).Delete(resourceName, &metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("Error while destroying Astarte Resource.")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Your Astarte instance has been successfully destroyed. Please allow a few minutes for the Cluster to settle and for all resource to be evicted.")

	if deleteVolumes {
		fmt.Println("Deleting all Volumes...")
		pvcList, err := kubernetesClient.CoreV1().PersistentVolumeClaims(resourceNamespace).List(metav1.ListOptions{})
		if err != nil {
			fmt.Println("Could not list PVCs. You might need to delete Volumes manually.")
			os.Exit(1)
		}

		for _, pvc := range pvcList.Items {
			// Check for all services which spawn a PVC, and delete it in case.
			for _, svc := range []string{"cfssl", "cassandra", "rabbitmq", "vernemq"} {
				if strings.HasPrefix(pvc.Name, resourceName+"-"+svc+"-data") {
					err = kubernetesClient.CoreV1().PersistentVolumeClaims(resourceNamespace).Delete(pvc.Name, &metav1.DeleteOptions{})
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
