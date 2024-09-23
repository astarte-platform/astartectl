// Copyright 2024 SECO Mind Srl
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

package realm

import (
	"encoding/json"
	"fmt"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	"os"
)

// triggersPoliciesCmd represents the triggers command
var triggersPoliciesCmd = &cobra.Command{
	Use:     "trigger-policies",
	Short:   "Manage trigger delivery policies",
	Long:    `List, show, install or delete trigger delivery policies in your realm.`,
	Aliases: []string{"trigger-policy"},
}

var triggersPoliciesListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List trigger policies",
	Long:    `List the name of trigger policies installed in the realm.`,
	Example: `  astartectl realm-management trigger-policies list`,
	RunE:    triggersPoliciesListtF,
	Aliases: []string{"ls"},
}

var triggersPoliciesShowCmd = &cobra.Command{
	Use:     "show <trigger_policy_name>",
	Short:   "Show trigger policy",
	Long:    `Shows a trigger policy installed in the realm.`,
	Example: `  astartectl realm-management trigger-policies show my_trigger_policiy`,
	Args:    cobra.ExactArgs(1),
	RunE:    triggersPoliciesShowF,
}

var triggersPoliciesInstallCmd = &cobra.Command{
	Use:   "install <trigger_policy_file>",
	Short: "Install trigger policy",
	Long: `Install the given trigger policy in the realm.
<trigger_file> must be a path to a JSON file containing a valid Astarte trigger policy.`,
	Example: `  astartectl realm-management trigger-policies install my_policy.json`,
	Args:    cobra.ExactArgs(1),
	RunE:    triggersPoliciesInstallF,
}

var triggersPoliciesDeleteCmd = &cobra.Command{
	Use:     "delete <trigger_policy_name>",
	Short:   "Delete a trigger policy",
	Long:    `Deletes the specified trigger policy from the realm.`,
	Example: `  astartectl realm-management trigger-policies delete my_trigger_policiy`,
	Args:    cobra.ExactArgs(1),
	RunE:    triggersPoliciesDeleteF,
	Aliases: []string{"del"},
}

func init() {
	RealmManagementCmd.AddCommand(triggersPoliciesCmd)
	triggersPoliciesCmd.AddCommand(
		triggersPoliciesListCmd,
		triggersPoliciesShowCmd,
		triggersPoliciesInstallCmd,
		triggersPoliciesDeleteCmd,
	)
}

func triggersPoliciesListtF(command *cobra.Command, args []string) error {
	realmPolicies, _ := listPolicies(realm)
	fmt.Println(realmPolicies)
	return nil
}

func triggersPoliciesShowF(command *cobra.Command, args []string) error {
	policyName := args[0]
	policyDefinition, _ := getPolicyDefinition(realm, policyName)
	respJSON, _ := json.MarshalIndent(policyDefinition, "", "  ")
	fmt.Println(string(respJSON))
	return nil
}

func triggersPoliciesInstallF(command *cobra.Command, args []string) error {
	triggerFile, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	var triggerPolicy map[string]interface{}
	err = json.Unmarshal(triggerFile, &triggerPolicy)
	if err != nil {
		return err
	}

	err = installTriggerPolicy(realm, triggerPolicy)
	if err != nil {
		fmt.Println("Something went wrong, check error description below")
		fmt.Println(err)
		return err
	}
	fmt.Println("ok")
	return nil
}

func triggersPoliciesDeleteF(command *cobra.Command, args []string) error {
	policyName := args[0]
	deleteTriggerPolicyCall, err := astarteAPIClient.DeleteTriggerDeliveryPolicy(realm, policyName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	utils.MaybeCurlAndExit(deleteTriggerPolicyCall, astarteAPIClient)

	deleteTriggerPolicyRes, err := deleteTriggerPolicyCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, _ = deleteTriggerPolicyRes.Parse()

	fmt.Println("ok")
	return nil
}

func installTriggerPolicy(realm string, policy any) error {
	installTriggerDeliveryCall, err := astarteAPIClient.InstallTriggerDeliveryPolicy(realm, policy)
	if err != nil {
		return err
	}

	utils.MaybeCurlAndExit(installTriggerDeliveryCall, astarteAPIClient)

	installTriggerDeliveryRes, err := installTriggerDeliveryCall.Run(astarteAPIClient)
	if err != nil {
		return err
	}
	_, _ = installTriggerDeliveryRes.Parse()
	return nil
}

func listPolicies(realm string) ([]string, error) {
	ListTriggerPolicyDeliveryCall, err := astarteAPIClient.ListTriggerDeliveryPolicies(realm)
	if err != nil {
		return []string{}, err
	}

	utils.MaybeCurlAndExit(ListTriggerPolicyDeliveryCall, astarteAPIClient)

	ListTriggerPolicyRes, err := ListTriggerPolicyDeliveryCall.Run(astarteAPIClient)
	if err != nil {
		return []string{}, err
	}
	rawlistPolicies, err := ListTriggerPolicyRes.Parse()
	if err != nil {
		return []string{}, err
	}
	return rawlistPolicies.([]string), nil
}

func getPolicyDefinition(realm, policyName string) (map[string]interface{}, error) {
	getPolicyCall, err := astarteAPIClient.GetTriggerDeliveryPolicy(realm, policyName)
	if err != nil {
		return nil, err
	}

	utils.MaybeCurlAndExit(getPolicyCall, astarteAPIClient)

	getPolicyrRes, err := getPolicyCall.Run(astarteAPIClient)
	if err != nil {
		return nil, err
	}
	rawPolicy, err := getPolicyrRes.Parse()
	if err != nil {
		return nil, err
	}

	return rawPolicy.(map[string]interface{}), nil
}
