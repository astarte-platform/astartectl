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

	"github.com/Masterminds/semver/v3"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
)

var upgradeOperatorCmd = &cobra.Command{
	Use:   "upgrade-operator",
	Short: "Upgrade Astarte Operator in the current Kubernetes Cluster",
	Long: `Upgrade Astarte Operator in the current Kubernetes Cluster. This will adhere to the same current-context
kubectl mentions. If no versions are specified, the last stable version is used as the upgrade target..`,
	Example: `  astartectl cluster upgrade-operator`,
	RunE:    clusterUpgradeOperatorF,
}

func init() {
	upgradeOperatorCmd.PersistentFlags().String("version", "", "Version of Astarte Operator to upgrade to. If not specified, last stable version will be installed (recommended)")
	upgradeOperatorCmd.PersistentFlags().BoolP("non-interactive", "y", false, "Non-interactive mode. Will answer yes by default to all questions.")

	ClusterCmd.AddCommand(upgradeOperatorCmd)
}

func clusterUpgradeOperatorF(command *cobra.Command, args []string) error {
	currentAstarteOperator, err := getAstarteOperator()
	if err != nil {
		fmt.Println("Astarte Operator is not installed in your cluster. You probably want to use astartectl cluster install-operator.")
		os.Exit(1)
	}
	currentAstarteOperatorVersion, err := semver.NewVersion(strings.Split(currentAstarteOperator.Spec.Template.Spec.Containers[0].Image, ":")[1])
	if err != nil {
		return err
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
			fmt.Println(err)
			os.Exit(1)
		}
	}
	upgradeVersion, err := semver.NewVersion(version)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !upgradeVersion.GreaterThan(currentAstarteOperatorVersion) {
		fmt.Printf("You're currently running Astarte Operator version %s, no updates are available.\n", currentAstarteOperatorVersion)
		return nil
	}
	fmt.Printf("Will upgrade Astarte Operator to version %s.\n", version)

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

	// This section for now it's basically the same as we just need to upgrade all the resources. Moving forward, should the
	// Operator change drammatically, we'll need proper cleanups+upgrades depending on the Operator version.

	fmt.Println("Upgrading RBAC Roles...")

	// Service Account
	serviceAccount := unmarshalYAML("deploy/service_account.yaml", version)
	_, err = kubernetesClient.CoreV1().ServiceAccounts("kube-system").Update(serviceAccount.(*corev1.ServiceAccount))
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("WARNING: Service Account already exists in the cluster.")
		} else {
			fmt.Println("Error while deploying Service Account. Your deployment might be incomplete.")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Cluster Role
	role := unmarshalYAML("deploy/role.yaml", version)
	_, err = kubernetesClient.RbacV1().ClusterRoles().Update(role.(*rbacv1.ClusterRole))
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("WARNING: Cluster Role already exists in the cluster.")
		} else {
			fmt.Println("Error while deploying Service Account. Your deployment might be incomplete.")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Cluster Role Binding
	roleBinding := unmarshalYAML("deploy/role_binding.yaml", version)
	_, err = kubernetesClient.RbacV1().ClusterRoleBindings().Update(roleBinding.(*rbacv1.ClusterRoleBinding))
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("WARNING: Cluster Role Binding already exists in the cluster.")
		} else {
			fmt.Println("Error while deploying Service Account. Your deployment might be incomplete.")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Println("RBAC Roles Successfully upgraded.")
	fmt.Println("Upgrading Astarte Custom Resource Definitions...")

	// This is where it gets tricky. For all supported CRDs, we need to either update or install them. When we update,
	// we need to ensure that the resourceVersion is increased compared to the existing resource.

	err = upgradeCRD("deploy/crds/api_v1alpha1_astarte_crd.yaml", version, currentAstarteOperatorVersion.Original())
	if err != nil {
		err = upgradeCRD("deploy/crds/api_v1alpha1_astarte_voyager_ingress_crd.yaml", version, currentAstarteOperatorVersion.Original())
	}
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("WARNING: AstarteVoyagerIngress CRD already exists in the cluster.")
		} else {
			fmt.Println("Error while deploying AstarteVoyagerIngress CRD. Your deployment might be incomplete.")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Println("Astarte Custom Resource Definitions successfully upgraded.")
	fmt.Println("Upgrading Astarte Operator...")

	// Astarte Operator Deployment
	astarteOperator := unmarshalYAML("deploy/operator.yaml", version)
	astarteOperatorDeployment, err := kubernetesClient.AppsV1().Deployments("kube-system").Update(astarteOperator.(*appsv1.Deployment))
	if err != nil {
		fmt.Println("Error while deploying Astarte Operator Deployment. Your deployment might be incomplete.")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Astarte Operator successfully upgraded. Waiting until it is ready...")

	var timeoutSeconds int64 = 60
	watcher, err := kubernetesClient.AppsV1().Deployments("kube-system").Watch(metav1.ListOptions{TimeoutSeconds: &timeoutSeconds})
	if err != nil {
		fmt.Println("Could not watch the Deployment state. However, deployment might be complete. Check with astartectl cluster show in a while.")
		fmt.Println(err)
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
			fmt.Println("Astarte Operator deployment ready! Check the state of your cluster with astartectl cluster show. Note that you might need to upgrade some of your Astarte instances depending on your Operator version.")
			return nil
		}
	}

	fmt.Println("Could not verify if Astarte Operator Deployment was successful. Please check the state of your cluster with astartectl cluster show.")
	os.Exit(1)
	return nil
}

func upgradeCRD(path, version, originalVersion string) error {
	// TODO: Handle v1, when we start planning on supporting it.
	crd := unmarshalYAML(path, version)
	currentCRD, err := kubernetesDynamicClient.Resource(crdResource).Get(
		crd.(*apiextensionsv1beta1.CustomResourceDefinition).Name, metav1.GetOptions{})
	if err != nil || currentCRD == nil {
		// It does not exist - go ahead and install it.
		_, err = kubernetesAPIExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(
			crd.(*apiextensionsv1beta1.CustomResourceDefinition))
		if err != nil {
			return err
		}
	} else {
		// Move to a 3-way JSON Merge patch
		originalCRD := unmarshalYAML(path, originalVersion)
		crdJSON, err := runtimeObjectToJSON(crd)
		currentCRDJSON, err := runtimeObjectToJSON(currentCRD)
		originalCRDJSON, err := runtimeObjectToJSON(originalCRD)

		preconditions := []mergepatch.PreconditionFunc{mergepatch.RequireKeyUnchanged("apiVersion"),
			mergepatch.RequireKeyUnchanged("kind"), mergepatch.RequireMetadataKeyUnchanged("name")}
		patch, err := jsonmergepatch.CreateThreeWayJSONMergePatch(originalCRDJSON, crdJSON, currentCRDJSON,
			preconditions...)

		_, err = kubernetesAPIExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Patch(
			crd.(*apiextensionsv1beta1.CustomResourceDefinition).Name, types.MergePatchType, patch)
		if err != nil {
			return err
		}
	}

	// All good.
	return nil
}
