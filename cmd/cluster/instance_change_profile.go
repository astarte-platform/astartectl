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
	"encoding/json"
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/astarte-platform/astartectl/cmd/cluster/deployment"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
)

var instanceChangeProfileCmd = &cobra.Command{
	Use:   "change-profile <name> [profile]",
	Short: "Changes the profile of an existing Astarte Instance",
	Long: `Changes the profile of an existing Astarte Instance in the current Kubernetes Cluster. If profile isn't specified,
astartectl will prompt the user with a set of available profiles which can be used.`,
	Example: `  astartectl cluster instances change-profile astarte basic`,
	RunE:    instanceChangeProfileF,
	Args:    cobra.RangeArgs(1, 2),
}

func init() {
	instanceChangeProfileCmd.PersistentFlags().StringP("namespace", "n", "astarte", "Namespace in which to look for the Astarte resource.")
	instanceChangeProfileCmd.PersistentFlags().BoolP("non-interactive", "y", false, "Non-interactive mode. Will answer yes by default to all questions.")

	InstancesCmd.AddCommand(instanceChangeProfileCmd)
}

func instanceChangeProfileF(command *cobra.Command, args []string) error {
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

	astarteSpec := astarteObject.Object["spec"].(map[string]interface{})
	_, deploymentManager, deploymentProfile := getManagedAstarteResourceStatus(*astarteObject)
	oldAstarteVersion, err := semver.NewVersion(astarteSpec["version"].(string))
	if err != nil {
		fmt.Printf("Installed version %s is not a valid Astarte version. Please ensure your Astarte installation is manageable by astartectl.", astarteSpec["version"].(string))
		os.Exit(1)
	}

	if deploymentManager != "astartectl" {
		fmt.Println("WARNING: It looks like this Astarte deployment isn't managed by astartectl. On paper, everything should still work, but have extra care in reviewing changes once done.")
	}

	newProfile := ""
	if len(args) == 2 {
		newProfile = args[1]
	}
	astarteDeployment := deployment.AstarteClusterProfile{}

	if deploymentProfile == "" {
		fmt.Println("Your Astarte instance has no profile associated. I will try to reconcile you with the profile you're going to choose.")
	} else if newProfile == "" {
		fmt.Printf("Your Astarte instance is running on '%s'. Let me inspect your cluster to find out if there are any other profiles you can choose from.\n", deploymentProfile)
	} else if newProfile == deploymentProfile {
		fmt.Printf("Your Astarte instance is already running on '%s'. All is good!\n", deploymentProfile)
		os.Exit(0)
	}

	if newProfile == "" {
		// Get the profile
		newProfile, astarteDeployment, err = promptForProfileExcluding(command, oldAstarteVersion, []string{deploymentProfile})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		// This is much easier. Just match the profile with either its corresponding profile for the next release, or
		// just ensure that nothing has changed in the profile itself. Should that be the case, we just bump Astarte's version
		// and call it a day.
		astarteDeployment = deployment.GetMatchingProfile(newProfile, oldAstarteVersion)
	}

	newResource := map[string]interface{}{}
	if astarteDeployment.IsValid() {
		// Ok. Build it with the standard mechanism.
		newResource = createAstarteResourceFromExistingSpecOrDie(command, resourceName, resourceNamespace, oldAstarteVersion,
			newProfile, astarteDeployment, astarteSpec)
	} else {
		// Fail.
		fmt.Printf("I found no matching '%s' profile for Astarte %s. Maybe you spelled the profile wrong, or you should upgrade astartectl first?\n", newProfile, oldAstarteVersion)
		os.Exit(1)
	}

	// Time for some Kubernetes dark magic
	preconditions := []mergepatch.PreconditionFunc{mergepatch.RequireMetadataKeyUnchanged("namespace"), mergepatch.RequireMetadataKeyUnchanged("name")}
	oldResourceJSON, err := runtimeObjectToJSON(astarteObject)
	newResourceJSON, err := json.Marshal(newResource)

	// To build the original resource, kill the status field, and replace all metadata with the new metadata.
	originalResourceBytes, err := runtimeObjectToJSON(astarteObject)
	originalResourceJSON := map[string]interface{}{}
	json.Unmarshal(originalResourceBytes, &originalResourceJSON)
	delete(originalResourceJSON, "status")
	delete(originalResourceJSON, "metadata")
	originalResourceJSON["metadata"] = newResource["metadata"]
	originalResourceBytes, err = json.Marshal(originalResourceJSON)

	// Build the patch
	patch, err := jsonmergepatch.CreateThreeWayJSONMergePatch(originalResourceBytes, newResourceJSON, oldResourceJSON,
		preconditions...)

	review, err := utils.AskForConfirmation("Ready to change profile! Would you like to review the patch?")
	if review {
		fmt.Println(string(patch))
	}
	goAhead, err := utils.AskForConfirmation(fmt.Sprintf("Your Astarte instance profile will be switched from '%v' to '%v'. Would you like to continue?",
		deploymentProfile, newProfile))
	if !goAhead {
		fmt.Println("Oh, okay :(")
		os.Exit(0)
	}

	fmt.Println("Ok. Hold on a second while I change your Astarte profile...")

	_, err = kubernetesDynamicClient.Resource(astarteV1Alpha1).Namespace(resourceNamespace).Patch(
		resourceName, types.MergePatchType, patch, v1.PatchOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Done! Astarte Operator will take over from here. Watch your cluster and monitor your resources to ensure the profile change was successful.")

	return nil
}
