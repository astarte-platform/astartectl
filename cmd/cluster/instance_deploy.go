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
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a minimal Astarte Instance in the current Kubernetes Cluster. Only for testing.",
	Long: `Deploy a minimal Astarte Instance in the current Kubernetes Cluster. This will adhere to the same current-context
kubectl mentions. If no versions are specified, the last stable version is deployed. This should only be used for testing purposes.`,
	Example: `  astartectl cluster instances deploy`,
	RunE:    clusterDeployF,
}

func init() {
	deployCmd.PersistentFlags().String("name", "", "Name of the deployed Astarte resource.")
	deployCmd.PersistentFlags().String("version", "", "Version of Astarte to deploy. If not specified, last stable version will be deployed.")
	deployCmd.PersistentFlags().String("api-host", "", "The API host for this Astarte deployment. If not specified, it will be prompted when deploying.")
	deployCmd.PersistentFlags().String("broker-host", "", "The Broker host for this Astarte deployment. If not specified, it will be prompted when deploying.")
	deployCmd.PersistentFlags().Int("broker-port", 8883, "The Broker port for this Astarte deployment. Defaults to 8883.")
	deployCmd.PersistentFlags().String("broker-tls-secret", "", "The existing TLS Secret, if any, to be used by the broker as its SSL Certificate. When specified, VerneMQ SSL listener will be enabled.")
	deployCmd.PersistentFlags().String("cassandra-nodes", "", "The Cassandra nodes the Astarte deployment should use for connecting. Valid only if the deployment profile has an external Cassandra.")
	deployCmd.PersistentFlags().String("cassandra-volume-size", "", "The Cassandra PVC size for this Astarte deployment. If not specified, it will be prompted when deploying.")
	deployCmd.PersistentFlags().String("cfssl-volume-size", "", "If Astarte is < 1.0.0, the CFSSL PVC size for this Astarte deployment. If not specified, it will be prompted when deploying.")
	deployCmd.PersistentFlags().String("cfssl-db-driver", "", "The CFSSL Database Driver. If not specified, it will default to SQLite.")
	deployCmd.PersistentFlags().String("cfssl-db-datasource", "", "The CFSSL Database Datasource. Compulsory when specifying a DB Driver different from SQLite.")
	deployCmd.PersistentFlags().String("rabbitmq-volume-size", "", "The RabbitMQ PVC size for this Astarte deployment. If not specified, it will be prompted when deploying.")
	deployCmd.PersistentFlags().String("vernemq-volume-size", "", "The VerneMQ PVC size for this Astarte deployment. If not specified, it will be prompted when deploying.")
	deployCmd.PersistentFlags().String("storage-class-name", "", "The Kubernetes Storage Class name for this Astarte deployment. If not specified, it will be left empty and the default Storage Class for your Cloud Provider will be used. Keep in mind that with some Cloud Providers, you always need to specify this.")
	deployCmd.PersistentFlags().Bool("no-ssl", false, "Don't use SSL for the API and Broker endpoints. Strongly not recommended.")
	deployCmd.PersistentFlags().BoolP("non-interactive", "y", false, "Non-interactive mode. Will answer yes by default to all questions.")

	InstancesCmd.AddCommand(deployCmd)
}

func clusterDeployF(command *cobra.Command, args []string) error {
	y, err := command.Flags().GetBool("non-interactive")
	if err != nil {
		return err
	}
	version, err := command.Flags().GetString("version")
	if err != nil {
		return err
	}
	if version == "" {
		latestAstarteVersion, _ := getLastAstarteRelease()
		version, err = utils.PromptChoice("What Astarte version would you like to install?", latestAstarteVersion, false, y)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	astarteVersion, err := semver.NewVersion(version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s is not a valid Astarte version", version)
		os.Exit(1)
	}

	profile, astarteDeployment, err := getBasicProfile(command, astarteVersion)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Create the Astarte Resource
	astarteDeploymentResource := createAstarteResourceOrDie(command, astarteVersion, profile, astarteDeployment)
	resourceName := astarteDeploymentResource["metadata"].(map[string]interface{})["name"].(string)
	resourceNamespace := astarteDeploymentResource["metadata"].(map[string]interface{})["namespace"].(string)

	//
	fmt.Println()
	fmt.Println("Your Astarte instance is ready to be deployed!")
	marshaledResource, err := yaml.Marshal(astarteDeploymentResource)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not build the YAML representation. Aborting.")
		os.Exit(1)
	}
	if !y {
		reviewConfiguration, _ := utils.AskForConfirmation("Do you wish to review the configuration before deployment?")
		if reviewConfiguration {
			fmt.Println(string(marshaledResource))
		}
		goAhead, _ := utils.AskForConfirmation(fmt.Sprintf("Your Astarte instance \"%s\" will be deployed in namespace \"%s\". Do you want to continue?", resourceName, resourceNamespace))
		if !goAhead {
			fmt.Println("Aborting.")
			os.Exit(0)
		}
	} else {
		fmt.Println(string(marshaledResource))
	}

	// Let's do it. Retrieve the namespace first and ensure it's there
	namespaceList, err := kubernetesClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	namespaceFound := false
	for _, ns := range namespaceList.Items {
		if ns.Name == resourceNamespace {
			namespaceFound = true
			break
		}
	}

	if !namespaceFound {
		fmt.Fprintf(os.Stderr, "Namespace %s does not exist, creating it...\n", resourceNamespace)
		nsSpec := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: resourceNamespace}}
		_, err := kubernetesClient.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metav1.CreateOptions{})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not create namespace!")
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	_, err = kubernetesDynamicClient.Resource(astarteV1Alpha1).Namespace(resourceNamespace).Create(
		context.TODO(), &unstructured.Unstructured{Object: astarteDeploymentResource}, metav1.CreateOptions{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while deploying Astarte Resource.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Your Astarte instance has been successfully deployed. Please allow a few minutes for the Cluster to start. You can monitor the progress with astartectl cluster show.")
	fmt.Println("Now waiting for Housekeeping setup to set up a context...")

	// 2 minute timeout
	for i := 0; i < 24; i++ {
		// Try every 5 seconds
		time.Sleep(5 * time.Second)
		if _, err = getHousekeepingKey(resourceName, resourceNamespace, false); err == nil {
			// Delegate this to the get-cluster-config implementation
			return doGetClusterConfig(resourceName, resourceNamespace)
		}
	}

	// We timed out. However, don't fail
	fmt.Fprintln(os.Stderr, "Could not fetch Housekeeping key! A context wasn't created, and you should check your deployment - something might be off")
	return nil
}
