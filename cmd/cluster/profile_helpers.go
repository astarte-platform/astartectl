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

	"github.com/Masterminds/semver/v3"
	"github.com/astarte-platform/astartectl/cmd/cluster/deployment"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func createAstarteResourceFromExistingSpecOrDie(command *cobra.Command, resourceName, resourceNamespace string, astarteVersion *semver.Version, profileName string, astarteDeployment deployment.AstarteClusterProfile, spec map[string]interface{}) map[string]interface{} {
	// Let's go
	astarteDeployment.DefaultSpec.Version = astarteVersion.String()
	astarteDeployment.DefaultSpec.API.Host = getStringFromSpecOrFlagOrPromptOrDie(spec, "api.host", command, "api-host", "Please enter the API Host for this Deployment:", "", false)
	astarteDeployment.DefaultSpec.Vernemq.Host = getStringFromSpecOrFlagOrPromptOrDie(spec, "vernemq.host", command, "broker-host", "Please enter the MQTT Broker Host for this Deployment:", "", false)
	storageClass := getStringFromSpecOrFlag(spec, "storageClassName", command, "storage-class-name")
	if storageClass != "" {
		astarteDeployment.DefaultSpec.StorageClassName = getStringFromSpecOrFlag(spec, "storageClassName", command, "storage-class-name")
	}

	// Ensure Storage and dependencies for all components.
	if astarteDeployment.DefaultSpec.Cassandra.Deploy {
		astarteDeployment.DefaultSpec.Cassandra.Storage.Size = getStringFromSpecOrFlagOrPromptOrDie(spec, "cassandra.storage.size", command, "cassandra-volume-size", "Please enter the Cassandra Volume size for this Deployment:",
			astarteDeployment.DefaultSpec.Cassandra.Storage.Size, false)
	} else {
		// Ask for nodes
		astarteDeployment.DefaultSpec.Cassandra.Nodes = getStringFromSpecOrFlagOrPromptOrDie(spec, "cassandra.nodes", command, "cassandra-nodes", "Please enter a comma separated list of Cassandra Nodes the cluster will connect to:",
			astarteDeployment.DefaultSpec.Cassandra.Nodes, false)
	}
	if astarteDeployment.DefaultSpec.Rabbitmq.Deploy {
		astarteDeployment.DefaultSpec.Rabbitmq.Storage.Size = getStringFromSpecOrFlagOrPromptOrDie(spec, "rabbitmq.storage.size", command, "rabbitmq-volume-size", "Please enter the RabbitMQ Volume size for this Deployment:",
			astarteDeployment.DefaultSpec.Rabbitmq.Storage.Size, false)
	}
	if astarteDeployment.DefaultSpec.Vernemq.Deploy {
		astarteDeployment.DefaultSpec.Vernemq.Storage.Size = getStringFromSpecOrFlagOrPromptOrDie(spec, "vernemq.storage.size", command, "vernemq-volume-size", "Please enter the VerneMQ Volume size for this Deployment:",
			astarteDeployment.DefaultSpec.Vernemq.Storage.Size, false)
	}
	if astarteDeployment.DefaultSpec.Cfssl.Deploy {
		astarteDeployment.DefaultSpec.Cfssl.Storage.Size = getStringFromSpecOrFlagOrPromptOrDie(spec, "cfssl.storage.size", command, "cfssl-volume-size", "Please enter the CFSSL Volume size for this Deployment:",
			astarteDeployment.DefaultSpec.Cfssl.Storage.Size, false)
		cfsslDBDriver := getStringFromSpecOrFlagOrPromptOrDie(spec, "cfssl.dbConfig.driver", command, "cfssl-db-driver", "Please enter the CFSSL DB Driver for this deployment.\nPlease note that leaving this empty will default to using SQLite, which is strongly discouraged in production.\nCFSSL DB Driver:",
			"", true)
		if cfsslDBDriver != "" && cfsslDBDriver != "sqlite3" {
			fmt.Println(cfsslDBDriver)
			astarteDeployment.DefaultSpec.Cfssl.DbConfig.Driver = cfsslDBDriver
			astarteDeployment.DefaultSpec.Cfssl.DbConfig.DataSource = getStringFromSpecOrFlagOrPromptOrDie(spec, "cfssl.dbConfig.dataSource", command, "cfssl-db-datasource", "Please enter the CFSSL DB Datasource (Connection URL) for this Deployment:",
				"", false)
		}
	}

	customFields := map[string]interface{}{}
	// Now we go with the custom fields
	for _, customizableField := range astarteDeployment.CustomizableFields {
		stringValue := getFromSpecOrPromptOrDie(spec, customizableField.Field, command, customizableField.Question,
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
		"astarte-platform.org/deployment-profile": profileName,
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

	return astarteDeploymentResource
}

func createAstarteResourceOrDie(command *cobra.Command, astarteVersion *semver.Version, profileName string, astarteDeployment deployment.AstarteClusterProfile) map[string]interface{} {
	resourceName := getStringFlagFromPromptOrDie(command, "name", "Please enter the name for this Astarte instance:", "astarte", false)
	resourceNamespace := getStringFlagFromPromptOrDie(command, "namespace", "Please enter the namespace where the Astarte instance will be deployed:", "astarte", false)
	// Reconciling with an empty spec will bear the very same effect
	return createAstarteResourceFromExistingSpecOrDie(command, resourceName, resourceNamespace, astarteVersion, profileName, astarteDeployment, map[string]interface{}{})
}
