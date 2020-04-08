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
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
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
	"k8s.io/apimachinery/pkg/runtime"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	apiextensions.Install(scheme.Scheme)
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

func runtimeObjectToJSON(object runtime.Object) ([]byte, error) {
	serializer := jsonserializer.NewSerializer(jsonserializer.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, false)
	buffer := bytes.NewBuffer([]byte{})
	err := serializer.Encode(object, buffer)
	return buffer.Bytes(), err
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

func getAstarte(astarteCRD dynamic.NamespaceableResourceInterface, name string, namespace string) (*unstructured.Unstructured, error) {
	return astarteCRD.Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func getAstarteOperator() (*appsv1.Deployment, error) {
	return kubernetesClient.AppsV1().Deployments("kube-system").Get(context.TODO(), "astarte-operator", metav1.GetOptions{})
}

func getLastOperatorRelease() (string, error) {
	return getLastReleaseForAstarteRepo("astarte-kubernetes-operator")
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

func getOperatorContent(path string, tag string) (string, error) {
	return getContentFromAstarteRepo("astarte-kubernetes-operator", path, tag)
}

func getContentFromAstarteRepo(repo string, path string, tag string) (string, error) {
	ctx := context.Background()
	client := github.NewClient(nil)

	ref := "v" + tag
	if strings.Contains(tag, "-snapshot") {
		// In this case, we want to fetch from the latest release branch.
		ref = "release-" + strings.Replace(tag, "-snapshot", "", -1)
	}

	content, _, _, err := client.Repositories.GetContents(ctx, "astarte-platform", repo,
		path, &github.RepositoryContentGetOptions{Ref: ref})

	if err != nil {
		return "", nil
	}

	return content.GetContent()
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

func getBaseVersionFromUnstable(version string) (string, error) {
	if !isUnstableVersion(version) {
		return "", fmt.Errorf("%v is not an unstable version", version)
	}

	if version == "snapshot" {
		return "", errors.New("You are running on snapshot - I have no way of reconciling you from here")
	}

	// Get the base version, and add a .0.
	baseVersion := strings.Replace(version, "-snapshot", "", 1)
	baseVersion += ".0"

	return baseVersion, nil
}

func promptForProfileExcluding(command *cobra.Command, astarteVersion *semver.Version, excluding []string) (string, deployment.AstarteClusterProfile, error) {
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
		return "", deployment.AstarteClusterProfile{}, fmt.Errorf("Unfortunately, your cluster allocatable resources do not allow for any profile to be deployed")
	}

	unexcludedAvailableProfiles := map[string]deployment.AstarteClusterProfile{}
	// Handle exclusion list
	for k, v := range availableProfiles {
		exclude := false
		for _, s := range excluding {
			if s == k {
				exclude = true
				break
			}
		}
		if exclude {
			continue
		}

		unexcludedAvailableProfiles[k] = v
	}

	if len(unexcludedAvailableProfiles) == 0 {
		return "", deployment.AstarteClusterProfile{}, fmt.Errorf("There are no other profiles which can be deployed on this cluster besides %s", excluding)
	}

	fmt.Println("You can safely deploy the following Profiles on this cluster:")
	for _, v := range unexcludedAvailableProfiles {
		fmt.Printf("%s: %s\n", v.Name, v.Description)
	}

	fmt.Println()
	profile := getStringFlagFromPromptOrDie(command, "profile", "Which profile would you like to deploy?", "", false)

	if _, ok := unexcludedAvailableProfiles[profile]; !ok {
		fmt.Printf("Profile %s does not exist! Aborting.\n", profile)
		os.Exit(1)
	}

	return profile, availableProfiles[profile], nil
}

func promptForProfile(command *cobra.Command, astarteVersion *semver.Version) (string, deployment.AstarteClusterProfile, error) {
	return promptForProfileExcluding(command, astarteVersion, []string{})
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
		fmt.Println(err)
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
			fmt.Println(err)
			os.Exit(1)
		}
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
