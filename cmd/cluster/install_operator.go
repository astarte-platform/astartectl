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

	"github.com/Masterminds/semver/v3"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var installCmd = &cobra.Command{
	Use:   "install-operator",
	Short: "Install Astarte Operator in the current Kubernetes Cluster",
	Long: `Install Astarte Operator in the current Kubernetes Cluster. This will adhere to the same current-context
kubectl mentions. If no versions are specified, the last stable version is installed.`,
	Example: `  astartectl cluster install-operator`,
	RunE:    clusterInstallF,
}

func init() {
	installCmd.PersistentFlags().String("version", "", "Version of Astarte Operator to install. If not specified, last stable version will be installed (recommended)")
	installCmd.PersistentFlags().BoolP("non-interactive", "y", false, "Non-interactive mode. Will answer yes by default to all questions.")

	ClusterCmd.AddCommand(installCmd)
}

func clusterInstallF(command *cobra.Command, args []string) error {
	_, err := getAstarteOperator()
	if err == nil {
		fmt.Println("Astarte Operator is already installed in your cluster.")
		os.Exit(1)
	}

	version, err := command.Flags().GetString("version")
	if err != nil {
		return err
	}
	nonInteractive, err := command.Flags().GetBool("non-interactive")
	if err != nil {
		return err
	}

	if version == "" {
		version, err = getLastOperatorRelease()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	semVersion, err := semver.NewVersion(version)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if isUnstableVersion(version) {
		fmt.Fprintln(os.Stderr, "You're trying install in your cluster an unstable snapshot - this is usually is a bad idea. Make sure you know what you're doing.")
	}

	fmt.Printf("Will install Astarte Operator version %s in the Cluster.\n", version)
	if !nonInteractive {
		confirmation, err := utils.AskForConfirmation("Do you want to continue?")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if !confirmation {
			return nil
		}
	}

	fmt.Println("Installing RBAC Roles...")

	// Service Account
	serviceAccount := unmarshalYAML("deploy/service_account.yaml", version)
	_, err = kubernetesClient.CoreV1().ServiceAccounts("kube-system").Create(
		context.TODO(), serviceAccount.(*corev1.ServiceAccount), metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Fprintln(os.Stderr, "WARNING: Service Account already exists in the cluster.")
		} else {
			fmt.Fprintln(os.Stderr, "Error while deploying Service Account. Your deployment might be incomplete.")
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	// Cluster Role
	role := unmarshalYAML("deploy/role.yaml", version)
	_, err = kubernetesClient.RbacV1().ClusterRoles().Create(context.TODO(), role.(*rbacv1.ClusterRole), metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Fprintln(os.Stderr, "WARNING: Cluster Role already exists in the cluster.")
		} else {
			fmt.Fprintln(os.Stderr, "Error while deploying Service Account. Your deployment might be incomplete.")
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	// Cluster Role Binding
	roleBinding := unmarshalYAML("deploy/role_binding.yaml", version)
	_, err = kubernetesClient.RbacV1().ClusterRoleBindings().Create(
		context.TODO(), roleBinding.(*rbacv1.ClusterRoleBinding), metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Fprintln(os.Stderr, "WARNING: Cluster Role Binding already exists in the cluster.")
		} else {
			fmt.Fprintln(os.Stderr, "Error while deploying Service Account. Your deployment might be incomplete.")
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	fmt.Println("RBAC Roles Successfully installed.")
	fmt.Println("Installing Astarte Custom Resource Definitions...")

	astarteCRDName := "api.astarte-platform.org_astartes_crd.yaml"
	aviCRDName := "api.astarte-platform.org_astartevoyageringresses_crd.yaml"
	c, _ := semver.NewConstraint("< 0.10.99")
	if c.Check(semVersion) {
		// Use old CRD filenames
		astarteCRDName = "api_v1alpha1_astarte_crd.yaml"
		aviCRDName = "api_v1alpha1_astarte_voyager_ingress_crd.yaml"
	}

	astarteCRD := unmarshalYAML("deploy/crds/"+astarteCRDName, version)
	_, err = kubernetesAPIExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(
		context.TODO(), astarteCRD.(*apiextensionsv1beta1.CustomResourceDefinition), metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Fprintln(os.Stderr, "WARNING: Astarte CRD already exists in the cluster.")
		} else {
			fmt.Fprintln(os.Stderr, "Error while deploying Astarte CRD. Your deployment might be incomplete.")
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	astarteVoyagerIngressCRD := unmarshalYAML("deploy/crds/"+aviCRDName, version)
	_, err = kubernetesAPIExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(
		context.TODO(), astarteVoyagerIngressCRD.(*apiextensionsv1beta1.CustomResourceDefinition), metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Fprintln(os.Stderr, "WARNING: AstarteVoyagerIngress CRD already exists in the cluster.")
		} else {
			fmt.Fprintln(os.Stderr, "Error while deploying AstarteVoyagerIngress CRD. Your deployment might be incomplete.")
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	fmt.Println("Installing Astarte Custom Resource Definitions successfully installed.")
	fmt.Println("Installing Astarte Operator...")

	// Astarte Operator Deployment
	astarteOperator := unmarshalYAML("deploy/operator.yaml", version)
	astarteOperatorDeployment, err := kubernetesClient.AppsV1().Deployments("kube-system").Create(
		context.TODO(), astarteOperator.(*appsv1.Deployment), metav1.CreateOptions{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while deploying Astarte Operator Deployment. Your deployment might be incomplete.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Astarte Operator successfully installed. Waiting until it is ready...")

	var timeoutSeconds int64 = 60
	watcher, err := kubernetesClient.AppsV1().Deployments("kube-system").Watch(
		context.TODO(), metav1.ListOptions{TimeoutSeconds: &timeoutSeconds})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not watch the Deployment state. However, deployment might be complete. Check with astartectl cluster show in a while.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ch := watcher.ResultChan()
	for {
		event := <-ch
		deployment, ok := event.Object.(*appsv1.Deployment)
		if !ok {
			break
		}
		if deployment.Name != astarteOperatorDeployment.GetObjectMeta().GetName() {
			continue
		}

		if deployment.Status.ReadyReplicas >= 1 {
			fmt.Println("Astarte Operator deployment ready! Check the state of your cluster with astartectl cluster show, and then deploy your Astarte installation with astartectl cluster instance deploy.")
			return nil
		}
	}

	fmt.Println("Could not verify if Astarte Operator Deployment was successful. Please check the state of your cluster with astartectl cluster show.")
	os.Exit(1)
	return nil
}
