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

	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall-operator",
	Short: "Uninstall Astarte Operator from the current Kubernetes Cluster",
	Long: `Uninstall Astarte Operator from the current Kubernetes Cluster. This will adhere to the same current-context
kubectl mentions. This command will refuse to run unless no Astarte instances are managed by this Cluster.`,
	Example: `  astartectl cluster uninstall-operator`,
	RunE:    clusterUninstallF,
}

func init() {
	uninstallCmd.PersistentFlags().BoolP("non-interactive", "y", false, "Non-interactive mode. Will answer yes by default to all questions.")

	ClusterCmd.AddCommand(uninstallCmd)
}

func clusterUninstallF(command *cobra.Command, args []string) error {
	_, err := getAstarteOperator()
	if err != nil {
		fmt.Println("Astarte Operator is not installed in your cluster.")
		os.Exit(1)
	}

	astartes, err := listAstartes()
	if err == nil && len(astartes) > 0 {
		fmt.Println("Your cluster has at least one active Astarte instance managed by the Operator. You can uninstall the operator only if no instances are deployed.")
		os.Exit(1)
	}

	nonInteractive, err := command.Flags().GetBool("non-interactive")
	if err != nil {
		return err
	}

	fmt.Println("Will uninstall Astarte Operator from the Cluster.")
	if !nonInteractive {
		confirmation, err := utils.AskForConfirmation("Do you want to continue?")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if !confirmation {
			return nil
		}
	}

	// Delete Operator Deployment
	err = kubernetesClient.AppsV1().Deployments("kube-system").Delete(
		context.TODO(), "astarte-operator", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("WARNING: Could not delete Astarte Operator Deployment.")
		fmt.Println(err)
	}

	// Delete Cluster Role Binding
	err = kubernetesClient.RbacV1().ClusterRoleBindings().Delete(
		context.TODO(), "astarte-operator", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("WARNING: Could not delete Cluster Role Binding.")
		fmt.Println(err)
	}

	// Delete Cluster Role
	err = kubernetesClient.RbacV1().ClusterRoles().Delete(
		context.TODO(), "astarte-operator", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("WARNING: Could not delete Cluster Role.")
		fmt.Println(err)
	}

	// Delete Service Account
	err = kubernetesClient.CoreV1().ServiceAccounts("kube-system").Delete(
		context.TODO(), "astarte-operator", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("WARNING: Could not delete Service Account.")
		fmt.Println(err)
	}

	// Delete Astarte CRD
	err = kubernetesAPIExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(
		context.TODO(), "astartes.api.astarte-platform.org", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("WARNING: Could not delete Astarte CRD.")
		fmt.Println(err)
	}

	// Delete AstarteVoyagerIngress CRD
	err = kubernetesAPIExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(
		context.TODO(), "astartevoyageringresses.api.astarte-platform.org", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("WARNING: Could not delete Astarte CRD.")
		fmt.Println(err)
	}

	fmt.Println("Astarte Operator has been successfully uninstalled from your cluster.")

	return nil
}
