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
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/astarte-platform/astartectl/cmd/cluster/deployment"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
)

var instanceUpgradeCmd = &cobra.Command{
	Use:     "upgrade <name> <version>",
	Short:   "Shows details about an Astarte Instance in the current Kubernetes Cluster",
	Long:    `Shows details about an Astarte Instance in the current Kubernetes Cluster.`,
	Example: `  astartectl cluster instances upgrade astarte 0.10.2`,
	RunE:    instanceUpgradeF,
	Args:    cobra.RangeArgs(1, 2),
}

func init() {
	instanceUpgradeCmd.PersistentFlags().String("namespace", "astarte", "Namespace in which to look for the Astarte resource.")
	instanceUpgradeCmd.PersistentFlags().String("profile", "", "Astarte Deployment Profile. Ignored if the existing Astarte instance is already associated to a Profile. If not specified and not associated yet, it will be prompted when deploying.")
	instanceUpgradeCmd.PersistentFlags().BoolP("non-interactive", "y", false, "Non-interactive mode. Will answer yes by default to all questions.")

	InstancesCmd.AddCommand(instanceUpgradeCmd)
}

func instanceUpgradeF(command *cobra.Command, args []string) error {
	resourceName := args[0]
	resourceNamespace, err := command.Flags().GetString("namespace")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if resourceNamespace == "" {
		resourceNamespace = "astarte"
	}

	astarteObject, err := getAstarteInstance(resourceName, resourceNamespace)
	if err != nil {
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

	// Good. Let's check the version now.
	version := ""
	if len(args) == 2 {
		version = args[1]
	}
	if version == "" {
		latestAstarteVersion, _ := getLastAstarteRelease()
		latestAstarteSemVersion, err := semver.NewVersion(latestAstarteVersion)
		if latestAstarteSemVersion.Compare(oldAstarteVersion) < 1 {
			fmt.Printf("Latest released Astarte version is %s. You're on latest and greatest!\n", latestAstarteVersion)
			os.Exit(0)
		}
		version, err = utils.PromptChoice("What Astarte version would you like to upgrade to?", latestAstarteVersion, false)
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

	if astarteVersion.Compare(oldAstarteVersion) < 1 {
		fmt.Printf("Your Astarte cluster is running Astarte %s, no need for upgrades today.", version)
		os.Exit(0)
	}

	astarteDeployment := deployment.AstarteClusterProfile{}

	if deploymentProfile == "" {
		fmt.Println("It looks like this deployment has no profile associated. To move forward, you should probably associate a profile. You may choose not to, but in that case, I won't be able to help you if anything changed in the Operator.")
		associateProfile, err := utils.AskForConfirmation("Would you like to associate a profile?")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if associateProfile {
			fmt.Println("Whew, good choice. Let me inspect your cluster and tell you what's available. Then, I'll do my best to ask you only for what's strictly needed:")
			fmt.Println()

			// Get the profile
			deploymentProfile, astarteDeployment, err = promptForProfile(command, astarteVersion)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Ok, I'll just try and upgrade blindly. Don't say I didn't warn you!")
		}
	} else {
		// This is much easier. Just match the profile with either its corresponding profile for the next release, or
		// just ensure that nothing has changed in the profile itself. Should that be the case, we just bump Astarte's version
		// and call it a day.
		astarteDeployment = deployment.GetMatchingProfile(deploymentProfile, astarteVersion)
		if !astarteDeployment.IsValid() {
			fmt.Printf("I found no matching '%s' profile for Astarte %s. Maybe upgrade astartectl?\n", deploymentProfile, version)
			os.Exit(1)
		}
		fmt.Println("Found a matching profile. Let me handle the hard stuff for you then, you're in good hands!")
	}

	newResource := map[string]interface{}{}
	if astarteDeployment.IsValid() {
		// Ok. Build it with the standard mechanism.
		newResource = createAstarteResourceFromExistingSpecOrDie(command, resourceName, resourceNamespace, astarteVersion,
			deploymentProfile, astarteDeployment, astarteSpec)
	} else {
		// Patch the hell out of the existing resource, and hope for the best.
		newResource = astarteObject.DeepCopy().Object
		newResource["spec"].(map[string]interface{})["version"] = version
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

	review, err := utils.AskForConfirmation("Ready to upgrade! Would you like to review the patch?")
	if review {
		fmt.Println(string(patch))
	}
	goAhead, err := utils.AskForConfirmation(fmt.Sprintf("Your cluster will be upgraded from version %v to version %v. Would you like to continue?",
		oldAstarteVersion.Original(), version))
	if !goAhead {
		fmt.Println("Oh, okay :(")
		os.Exit(0)
	}

	fmt.Println("Ok. Hold on a second while I upgrade Astarte...")

	_, err = kubernetesDynamicClient.Resource(astarteV1Alpha1).Namespace(resourceNamespace).Patch(
		context.TODO(), resourceName, types.MergePatchType, patch, v1.PatchOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Done! Astarte Operator will take over from here. Watch your cluster and monitor your resources to ensure the upgrade was successful.")

	return nil
}
