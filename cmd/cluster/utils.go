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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"code.cloudfoundry.org/bytefmt"
	"github.com/Masterminds/semver/v3"
	"github.com/astarte-platform/astartectl/cmd/cluster/deployment"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/google/go-github/v30/github"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/install"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/yaml"
)

func init() {
	apiextensions.Install(scheme.Scheme)
}

func listAstartes(namespace string) (map[string]*unstructured.UnstructuredList, error) {
	ret := make(map[string]*unstructured.UnstructuredList)
	for k, v := range astarteResourceClients {
		var list *unstructured.UnstructuredList
		var err error
		if namespace != "" {
			list, err = v.Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
		} else {
			list, err = v.List(context.TODO(), metav1.ListOptions{})
		}
		if err != nil {
			return nil, err
		}
		if len(list.Items) > 0 {
			ret[k] = list
		}
	}

	return ret, nil
}

func getAstarteInstance(name, namespace string) (*unstructured.Unstructured, error) {
	astartes, err := listAstartes(namespace)
	if err != nil || len(astartes) == 0 {
		return nil, errors.New("no managed astarte installations found")
	}

	for _, v := range astartes {
		for _, res := range v.Items {
			if res.Object["metadata"].(map[string]interface{})["name"] == name {
				return res.DeepCopy(), nil
			}
		}
	}

	return nil, errors.New("no such astarte instance found")
}

func getHousekeepingKey(name, namespace string, checkFirst bool) ([]byte, error) {
	if checkFirst {
		if _, err := getAstarteInstance(name, namespace); err != nil {
			return nil, err
		}
	}

	secret, err := kubernetesClient.CoreV1().Secrets(namespace).Get(
		context.TODO(), fmt.Sprintf("%s-housekeeping-private-key", name), v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret.Data["private-key"], nil
}

func getAstarteOperator(operatorName, operatorNamespace string) (*appsv1.Deployment, error) {
	return kubernetesClient.AppsV1().Deployments(operatorNamespace).Get(context.TODO(), operatorName, metav1.GetOptions{})
}

func getLastAstarteRelease() (string, error) {
	return getLastReleaseForAstarteRepo("astarte")
}

func getLastReleaseForAstarteRepo(repo string) (string, error) {
	ctx := context.Background()
	client := github.NewClient(nil)

	tags, _, err := client.Repositories.ListTags(ctx, "astarte-platform", repo, &github.ListOptions{})
	if err != nil {
		return "", err
	}

	collection := semver.Collection{}

	for _, tag := range tags {
		ver, err := semver.NewVersion(strings.Replace(tag.GetName(), "v", "", -1))
		if err != nil {
			continue
		}
		if ver.Prerelease() != "" {
			continue
		}

		collection = append(collection, ver)
	}

	sort.Sort(collection)

	return collection[len(collection)-1].Original(), nil
}

func getClusterAllocatableResources() (int, int64, int64, error) {
	// List Nodes
	list, err := kubernetesClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return 0, 0, 0, nil
	}

	var allocatableCPU int64 = 0
	var allocatableMemory int64 = 0
	for _, node := range list.Items {
		nodeAllocatableCPU := node.Status.Allocatable.Cpu().ScaledValue(resource.Milli)
		if nodeAllocatableCPU <= 0 {
			return 0, 0, 0, fmt.Errorf("Could not retrieve allocatable CPU for node %s", node.GetName())
		}
		allocatableCPU += nodeAllocatableCPU
		// Get Int64 directly, as the value is always returned in bytes.
		nodeAllocatableMemory, ok := node.Status.Allocatable.Memory().AsInt64()
		if !ok {
			return 0, 0, 0, fmt.Errorf("Could not retrieve allocatable Memory for node %s", node.GetName())
		}
		allocatableMemory += nodeAllocatableMemory
	}

	return len(list.Items), allocatableCPU, allocatableMemory, nil
}

func getManagedAstarteResourceStatus(res unstructured.Unstructured) (string, string, string) {
	var operatorStatus string = "Initializing"
	var deploymentManager string = ""
	var deploymentProfile string = ""
	if status, ok := res.Object["status"]; ok {
		statusMap := status.(map[string]interface{})
		if oldOperatorStatus, ok := statusMap["conditions"]; ok {
			operatorStatus = oldOperatorStatus.([]interface{})[0].(map[string]interface{})["type"].(string)
		} else {
			operatorStatus = statusMap["phase"].(string)
		}
	}
	if annotations, ok := res.Object["metadata"].(map[string]interface{})["annotations"]; ok {
		if dM, ok := annotations.(map[string]interface{})["astarte-platform.org/deployment-manager"]; ok {
			deploymentManager = dM.(string)
		}
		if dP, ok := annotations.(map[string]interface{})["astarte-platform.org/deployment-profile"]; ok {
			deploymentProfile = dP.(string)
		}
	}

	return operatorStatus, deploymentManager, deploymentProfile
}

func isUnstableVersion(version string) bool {
	return strings.HasSuffix(version, "-snapshot") || version == "snapshot"
}

func getProfile(command *cobra.Command, astarteVersion *semver.Version, burst bool) (string, deployment.AstarteClusterProfile, error) {
	nodes, allocatableCPU, allocatableMemory, err := getClusterAllocatableResources()
	if err != nil {
		return "", deployment.AstarteClusterProfile{}, err
	}

	fmt.Printf("Cluster has %v nodes\n", nodes)
	fmt.Printf("Allocatable CPU is %vm\n", allocatableCPU)
	fmt.Printf("Allocatable Memory is %v\n", bytefmt.ByteSize(uint64(allocatableMemory)))
	fmt.Println()

	clusterRequirements := deployment.AstarteProfileRequirements{
		CPUAllocation:    allocatableCPU,
		MemoryAllocation: allocatableMemory,
		MinNodes:         nodes,
		MaxNodes:         nodes,
	}
	availableProfiles := deployment.GetProfilesForVersionAndRequirements(astarteVersion, clusterRequirements)

	if len(availableProfiles) == 0 {
		return "", deployment.AstarteClusterProfile{}, fmt.Errorf("Unfortunately, your cluster allocatable resources do not allow for an Astarte instance to be deployed")
	}

	// Invariant: since burst requirements are a strict subset of basic requirements,
	// if a cluster allows for a basic profile, it also allows for a burst one.
	if burst {
		return "burst", availableProfiles["burst"], nil
	}

	// basic type profiles are the default
	return "basic", availableProfiles["basic"], nil
}

func getValueFromSpec(spec map[string]interface{}, field string) interface{} {
	fieldTokens := strings.Split(field, ".")
	aMap := spec

	for i, f := range fieldTokens {
		if _, ok := aMap[f]; ok {
			switch v := aMap[f].(type) {
			case map[string]interface{}:
				aMap = v
			case string:
				if i == len(fieldTokens)-1 {
					return v
				}
			}
		}
	}

	return nil
}

func getStringFromSpecOrFlag(spec map[string]interface{}, field string, command *cobra.Command, flagName string) string {
	ret := getValueFromSpec(spec, field)
	if ret != nil {
		return ret.(string)
	}

	flag := command.Flags().Lookup(flagName)
	if flag == nil {
		// If the flag isn't defined, assume it's not a big deal.
		return ""
	}

	fl, err := command.Flags().GetString(flagName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return fl
}

func getFromSpecOrPromptOrDie(spec map[string]interface{}, field string, command *cobra.Command, question string, defaultValue string, allowEmpty bool) string {
	ret := getValueFromSpec(spec, field)
	if ret != nil {
		return ret.(string)
	}

	return getFromPromptOrDie(command, question, defaultValue, allowEmpty)
}

func getIntFromSpecOrFlagOrPromptOrDie(spec map[string]interface{}, field string, command *cobra.Command, flagName string, question string, defaultValue int, allowEmpty bool) int {
	ret := getValueFromSpec(spec, field)
	if ret != nil {
		return ret.(int)
	}

	return getIntFlagFromPromptOrDie(command, flagName, question, defaultValue, allowEmpty)
}

func getIntFlagFromPromptOrDie(command *cobra.Command, flagName string, question string, defaultValue int, allowEmpty bool) int {
	var ret int
	flag := command.Flags().Lookup(flagName)
	if flag != nil {
		var err error
		ret, err = command.Flags().GetInt(flagName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	// Given we have no real way to check, let's prompt (or attempt to if the defaults match)
	if defaultValue == ret {
		i := getFromPromptOrDie(command, question, strconv.Itoa(defaultValue), allowEmpty)
		var err error
		if ret, err = strconv.Atoi(i); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	return ret
}

func getStringFromSpecOrFlagOrPromptOrDie(spec map[string]interface{}, field string, command *cobra.Command, flagName string, question string, defaultValue string, allowEmpty bool) string {
	ret := getValueFromSpec(spec, field)
	if ret != nil {
		return ret.(string)
	}

	return getStringFlagFromPromptOrDie(command, flagName, question, defaultValue, allowEmpty)
}

func getStringFlagFromPromptOrDie(command *cobra.Command, flagName string, question string, defaultValue string, allowEmpty bool) string {
	ret := ""
	flag := command.Flags().Lookup(flagName)
	if flag != nil {
		var err error
		ret, err = command.Flags().GetString(flagName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if ret == "" {
		ret = getFromPromptOrDie(command, question, defaultValue, allowEmpty)
	}
	return ret
}

func getFromPromptOrDie(command *cobra.Command, question string, defaultValue string, allowEmpty bool) string {
	y, err := command.Flags().GetBool("non-interactive")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ret, err := utils.PromptChoice(question, defaultValue, allowEmpty, y)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
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

func unstructuredToJSON(in *unstructured.Unstructured) ([]byte, error) {
	out, err := json.Marshal(in.Object)
	if err != nil {
		return []byte{}, err
	}
	return out, nil
}

func unstructuredToYAML(in *unstructured.Unstructured) ([]byte, error) {
	j, err := unstructuredToJSON(in)
	if err != nil {
		return []byte{}, err
	}
	out, err := yaml.JSONToYAML(j)
	if err != nil {
		return []byte{}, err
	}
	return out, nil
}

func dumpResourceToYAMLFile(in *unstructured.Unstructured, filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	y, err := unstructuredToYAML(in)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath, []byte(y), 0644); err != nil {
		return err
	}

	return nil
}
