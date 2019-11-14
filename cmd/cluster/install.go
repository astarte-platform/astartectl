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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Astarte Operator in the current Kubernetes Cluster",
	Long: `Install Astarte Operator in the current Kubernetes Cluster. This will adhere to the same current-context
kubectl mentions. If no versions are specified, the last stable version is installed.`,
	Example: `  astartectl cluster install`,
	RunE:    clusterInstallF,
}

func init() {
	installCmd.PersistentFlags().String("version", "", "Version of Astarte Operator to install. If not specified, last stable version will be installed (recommended)")
	installCmd.PersistentFlags().BoolP("non-interactive", "y", false, "Non-interactive mode. Will answer yes by default to all questions.")

	ClusterCmd.AddCommand(installCmd)
}

func unmarshalYAML(res string, version string) runtime.Object {
	content, err := getOperatorContent(res, version)
	if err != nil {
		fmt.Println("Error while parsing Kubernetes Resources. Your deployment might be incomplete.")
		fmt.Println(err)
		os.Exit(1)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(content), nil, nil)
	if err != nil {
		fmt.Println("Error while parsing Kubernetes Resources. Your deployment might be incomplete.")
		fmt.Println(err)
		os.Exit(1)
	}

	return obj
}

func unmarshalOperatorContentYAMLToJSON(res string, version string) map[string]interface{} {
	content, err := getOperatorContent(res, version)
	jsonStruct, err := utils.UnmarshalYAMLToJSON([]byte(content))
	if err != nil {
		fmt.Println("Error while parsing Kubernetes Resources. Your deployment might be incomplete.")
		fmt.Println(err)
		os.Exit(1)
	}
	return jsonStruct
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
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Printf("Will install Astarte Operator version %s in the Cluster.\n", version)
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

	fmt.Println("Installing RBAC Roles...")

	// Service Account
	serviceAccount := unmarshalYAML("deploy/service_account.yaml", version)
	_, err = kubernetesClient.CoreV1().ServiceAccounts("kube-system").Create(serviceAccount.(*corev1.ServiceAccount))
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
	_, err = kubernetesClient.RbacV1().ClusterRoles().Create(role.(*rbacv1.ClusterRole))
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
	_, err = kubernetesClient.RbacV1().ClusterRoleBindings().Create(roleBinding.(*rbacv1.ClusterRoleBinding))
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("WARNING: Cluster Role Binding already exists in the cluster.")
		} else {
			fmt.Println("Error while deploying Service Account. Your deployment might be incomplete.")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Println("RBAC Roles Successfully installed.")
	fmt.Println("Installing Astarte Custom Resource Definitions...")

	astarteCRD := unmarshalOperatorContentYAMLToJSON("deploy/crds/api_v1alpha1_astarte_crd.yaml", version)
	_, err = kubernetesDynamicClient.Resource(crdResource).Create(&unstructured.Unstructured{Object: astarteCRD},
		metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("WARNING: Astarte CRD already exists in the cluster.")
		} else {
			fmt.Println("Error while deploying Astarte CRD. Your deployment might be incomplete.")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	astarteVoyagerIngressCRD := unmarshalOperatorContentYAMLToJSON("deploy/crds/api_v1alpha1_astarte_voyager_ingress_crd.yaml", version)
	_, err = kubernetesDynamicClient.Resource(crdResource).Create(&unstructured.Unstructured{Object: astarteVoyagerIngressCRD},
		metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("WARNING: AstarteVoyagerIngress CRD already exists in the cluster.")
		} else {
			fmt.Println("Error while deploying AstarteVoyagerIngress CRD. Your deployment might be incomplete.")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Println("Installing Astarte Custom Resource Definitions successfully installed.")
	fmt.Println("Installing Astarte Operator...")

	// Astarte Operator Deployment
	astarteOperator := unmarshalYAML("deploy/operator.yaml", version)
	astarteOperatorDeployment, err := kubernetesClient.AppsV1().Deployments("kube-system").Create(astarteOperator.(*appsv1.Deployment))
	if err != nil {
		fmt.Println("Error while deploying Astarte Operator Deployment. Your deployment might be incomplete.")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Astarte Operator successfully installed. Waiting until it is ready...")

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
			fmt.Println("Astarte Operator deployment ready! Check the state of your cluster with astartectl cluster show, and then deploy your Astarte installation with astartectl cluster deploy.")
			return nil
		}
	}

	fmt.Println("Could not verify if Astarte Operator Deployment was successful. Please check the state of your cluster with astartectl cluster show.")
	os.Exit(1)
	return nil
}
