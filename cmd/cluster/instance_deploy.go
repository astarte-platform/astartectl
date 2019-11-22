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
	"go/types"
	"os"
	"strconv"
	"strings"

	"code.cloudfoundry.org/bytefmt"
	"github.com/Masterminds/semver/v3"
	"github.com/astarte-platform/astartectl/cmd/cluster/deployment"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy an Astarte Instance in the current Kubernetes Cluster",
	Long: `Deploy an Astarte Instance in the current Kubernetes Cluster. This will adhere to the same current-context
kubectl mentions. If no versions are specified, the last stable version is deployed.`,
	Example: `  astartectl cluster instances deploy`,
	RunE:    clusterDeployF,
}

func init() {
	deployCmd.PersistentFlags().String("name", "", "Name of the deployed Astarte resource.")
	deployCmd.PersistentFlags().String("namespace", "", "Namespace in which the Astarte resource will be deployed.")
	deployCmd.PersistentFlags().String("version", "", "Version of Astarte to deploy. If not specified, last stable version will be deployed.")
	deployCmd.PersistentFlags().String("profile", "", "Astarte Deployment Profile. If not specified, it will be prompted when deploying.")
	deployCmd.PersistentFlags().String("api-host", "", "The API host for this Astarte deployment. If not specified, it will be prompted when deploying.")
	deployCmd.PersistentFlags().String("broker-host", "", "The Broker host for this Astarte deployment. If not specified, it will be prompted when deploying.")
	deployCmd.PersistentFlags().String("cassandra-nodes", "", "The Cassandra nodes the Astarte deployment should use for connecting. Valid only if the deployment profile has an external Cassandra.")
	deployCmd.PersistentFlags().String("cassandra-volume-size", "", "The Cassandra PVC size for this Astarte deployment. If not specified, it will be prompted when deploying.")
	deployCmd.PersistentFlags().String("cfssl-volume-size", "", "The CFSSL PVC size for this Astarte deployment. If not specified, it will be prompted when deploying.")
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
	nodes, allocatableCPU, allocatableMemory, err := getClusterAllocatableResources()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Cluster has %v nodes\n", nodes)
	fmt.Printf("Allocatable CPU is %vm\n", allocatableCPU)
	fmt.Printf("Allocatable Memory is %v\n", bytefmt.ByteSize(uint64(allocatableMemory)))
	fmt.Println()

	version, err := command.Flags().GetString("version")
	if err != nil {
		return err
	}
	if version == "" {
		latestAstarteVersion, _ := getLastAstarteRelease()
		version, err = utils.PromptChoice("What Astarte version would you like to install?", latestAstarteVersion, false)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	astarteVersion, err := semver.NewVersion(version)
	if err != nil {
		fmt.Printf("%s is not a valid Astarte version", version)
		os.Exit(1)
	}

	clusterRequirements := deployment.AstarteProfileRequirements{
		CPUAllocation:    allocatableCPU,
		MemoryAllocation: allocatableMemory,
		MinNodes:         nodes,
		MaxNodes:         nodes,
	}
	availableProfiles := deployment.GetProfilesForVersionAndRequirements(astarteVersion, clusterRequirements)

	if len(availableProfiles) == 0 {
		fmt.Println("Unfortunately, your cluster allocatable resources do not allow for any profile to be deployed.")
		os.Exit(1)
	}

	fmt.Println("You can safely deploy the following Profiles on this cluster:")
	for _, v := range availableProfiles {
		fmt.Printf("%s: %s\n", v.Name, v.Description)
	}

	fmt.Println()
	profile := getStringFlagFromPromptOrDie(command, "profile", "Which profile would you like to deploy?", "", false)

	astarteDeployment := availableProfiles[profile]

	// Let's go
	resourceName := getStringFlagFromPromptOrDie(command, "name", "Please enter the name for this Astarte instance:", "astarte", false)
	resourceNamespace := getStringFlagFromPromptOrDie(command, "namespace", "Please enter the namespace where the Astarte instance will be deployed:", "astarte", false)
	astarteDeployment.DefaultSpec.Version = astarteVersion.String()
	astarteDeployment.DefaultSpec.API.Host = getStringFlagFromPromptOrDie(command, "api-host", "Please enter the API Host for this Deployment:", "", false)
	astarteDeployment.DefaultSpec.Vernemq.Host = getStringFlagFromPromptOrDie(command, "broker-host", "Please enter the MQTT Broker Host for this Deployment:", "", false)
	storageClassName, err := command.Flags().GetString("storage-class-name")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if storageClassName != "" {
		astarteDeployment.DefaultSpec.StorageClassName = storageClassName
	}

	// Ensure Storage and dependencies for all components.
	if astarteDeployment.DefaultSpec.Cassandra.Deploy {
		astarteDeployment.DefaultSpec.Cassandra.Storage.Size = getStringFlagFromPromptOrDie(command, "cassandra-volume-size", "Please enter the Cassandra Volume size for this Deployment:",
			astarteDeployment.DefaultSpec.Cassandra.Storage.Size, false)
	} else {
		// Ask for nodes
		astarteDeployment.DefaultSpec.Cassandra.Nodes = getStringFlagFromPromptOrDie(command, "cassandra-nodes", "Please enter a comma separated list of Cassandra Nodes the cluster will connect to:",
			astarteDeployment.DefaultSpec.Cassandra.Nodes, false)
	}
	if astarteDeployment.DefaultSpec.Rabbitmq.Deploy {
		astarteDeployment.DefaultSpec.Rabbitmq.Storage.Size = getStringFlagFromPromptOrDie(command, "rabbitmq-volume-size", "Please enter the RabbitMQ Volume size for this Deployment:",
			astarteDeployment.DefaultSpec.Rabbitmq.Storage.Size, false)
	}
	if astarteDeployment.DefaultSpec.Vernemq.Deploy {
		astarteDeployment.DefaultSpec.Vernemq.Storage.Size = getStringFlagFromPromptOrDie(command, "vernemq-volume-size", "Please enter the VerneMQ Volume size for this Deployment:",
			astarteDeployment.DefaultSpec.Vernemq.Storage.Size, false)
	}
	if astarteDeployment.DefaultSpec.Cfssl.Deploy {
		astarteDeployment.DefaultSpec.Cfssl.Storage.Size = getStringFlagFromPromptOrDie(command, "cfssl-volume-size", "Please enter the CFSSL Volume size for this Deployment:",
			astarteDeployment.DefaultSpec.Cfssl.Storage.Size, false)
		cfsslDBDriver := getStringFlagFromPromptOrDie(command, "cfssl-db-driver", "Please enter the CFSSL DB Driver for this deployment.\nPlease note that leaving this empty will default to using SQLite, which is strongly discouraged in production.\nCFSSL DB Connection String:",
			"", true)
		if cfsslDBDriver != "" {
			astarteDeployment.DefaultSpec.Cfssl.DbConfig.Driver = cfsslDBDriver
			astarteDeployment.DefaultSpec.Cfssl.DbConfig.DataSource = getStringFlagFromPromptOrDie(command, "cfssl-db-datasource", "Please enter the CFSSL DB Datasource for this Deployment:",
				"", false)
		}
	}

	customFields := map[string]interface{}{}
	// Now we go with the custom fields
	for _, customizableField := range astarteDeployment.CustomizableFields {
		stringValue := getFromPromptOrDie(command, customizableField.Question,
			fmt.Sprintf("%v", customizableField.Default), customizableField.AllowEmpty)
		switch customizableField.Type {
		case types.Int:
			i, err := strconv.Atoi(stringValue)
			if err != nil {
				fmt.Printf("%v is not a valid value for %v.\n", stringValue, customizableField.Field)
				os.Exit(1)
			}
			customFields[customizableField.Field] = i
		case types.Bool:
			b, err := strconv.ParseBool(stringValue)
			if err != nil {
				fmt.Printf("%v is not a valid value for %v.\n", stringValue, customizableField.Field)
				os.Exit(1)
			}
			customFields[customizableField.Field] = b
		default:
			customFields[customizableField.Field] = stringValue
		}
	}

	// Assemble the Astarte resource
	astarteK8sDeployment := deployment.GetBaseAstartev1alpha1Deployment()
	astarteK8sDeployment.Metadata.Name = resourceName
	astarteK8sDeployment.Metadata.Namespace = resourceNamespace
	astartectlAnnotations := map[string]string{
		"astarte-platform.org/deployment-manager": "astartectl",
		"astarte-platform.org/deployment-profile": profile,
	}
	astarteK8sDeployment.Metadata.Annotations = astartectlAnnotations
	astarteK8sDeployment.Spec = astarteDeployment.DefaultSpec

	astarteDeploymentYaml, err := yaml.Marshal(astarteK8sDeployment)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	astarteDeploymentResource, err := utils.UnmarshalYAMLToJSON(astarteDeploymentYaml)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Go with the custom fields
	for customField, customFieldValue := range customFields {
		fieldTokens := strings.Split(customField, ".")
		astarteDeploymentResource["spec"] = setInMapRecursively(astarteDeploymentResource["spec"].(map[string]interface{}),
			fieldTokens, customFieldValue)
	}

	//
	fmt.Println()
	fmt.Println("Your Astarte instance is ready to be deployed!")
	reviewConfiguration, _ := utils.AskForConfirmation("Do you wish to review the configuration before deployment?")
	if reviewConfiguration {
		marshaledResource, err := yaml.Marshal(astarteDeploymentResource)
		if err != nil {
			fmt.Println("Could not build the YAML representation. Aborting.")
			os.Exit(1)
		}
		fmt.Println(string(marshaledResource))
	}
	goAhead, _ := utils.AskForConfirmation(fmt.Sprintf("Your Astarte instance \"%s\" will be deployed in namespace \"%s\". Do you want to continue?", resourceName, resourceNamespace))
	if !goAhead {
		fmt.Println("Aborting.")
		os.Exit(0)
	}

	// Let's do it. Retrieve the namespace first and ensure it's there
	namespaceList, err := kubernetesClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		fmt.Println(err)
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
		fmt.Printf("Namespace %s does not exist, creating it...\n", resourceNamespace)
		nsSpec := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: resourceNamespace}}
		_, err := kubernetesClient.CoreV1().Namespaces().Create(nsSpec)
		if err != nil {
			fmt.Println("Could not create namespace!")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	_, err = kubernetesDynamicClient.Resource(astarteV1Alpha1).Namespace(resourceNamespace).Create(&unstructured.Unstructured{Object: astarteDeploymentResource},
		metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Error while deploying Astarte Resource.")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Your Astarte instance has been successfully deployed. Please allow a few minutes for the Cluster to start. You can monitor the progress with astartectl cluster show.")
	return nil
}

func getStringFlagFromPromptOrDie(command *cobra.Command, flagName string, question string, defaultValue string, allowEmpty bool) string {
	ret, err := command.Flags().GetString(flagName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if ret == "" {
		ret = getFromPromptOrDie(command, question, defaultValue, allowEmpty)
	}
	return ret
}

func getFromPromptOrDie(command *cobra.Command, question string, defaultValue string, allowEmpty bool) string {
	ret, err := utils.PromptChoice(question, defaultValue, allowEmpty)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return ret
}

func setInMapRecursively(aMap map[string]interface{}, tokens []string, customFieldValue interface{}) map[string]interface{} {
	if len(tokens) == 1 {
		aMap[tokens[0]] = customFieldValue
	} else {
		// Pop first element
		var token string
		token, tokens = tokens[0], tokens[1:]
		if _, ok := aMap[token]; ok {
			aMap[token] = setInMapRecursively(aMap[token].(map[string]interface{}), tokens, customFieldValue)
		} else {
			aMap[token] = setInMapRecursively(make(map[string]interface{}), tokens, customFieldValue)
		}
	}
	return aMap
}
