// Copyright © 2022 SECO Mind Srl
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
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"

	migrationutils "github.com/astarte-platform/astartectl/cmd/cr-migration"
	"github.com/astarte-platform/astartectl/cmd/cr-migration/v1alpha3tov2alpha1"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Interact with your Astarte instance for performing migration tasks",
}

var tov2alpha1 = &cobra.Command{
	Use:     "v2alpha1",
	Short:   "Migrate Astarte CRs to v2alpha1",
	RunE:    migrateToV2Alpha1,
	Example: `astartectl cluster instances migrate api v2alpha1`,
}

var MigrateApiCmd = &cobra.Command{
	Use:     "api",
	Short:   "Migrate Astarte API from one version to another",
	Example: `astartectl cluster instances migrate api`,
	Args:    cobra.NoArgs,
}

var replaceVoyagerCmd = &cobra.Command{
	Use:   "replace-voyager",
	Short: "Migrate the AstarteVoyagerIngress resource to an equivalent AstarteDefaultIngress",
	Long: `Migrate the AstarteVoyagerIngress resource to an equivalent AstarteDefaultIngress.

The user is required to interactively prompt information such as the TLS certificate names and the AstarteDefaultIngress resource name.
Additionally, for backup purposes, the --out flag allows to dump the AstarteVoyagerIngress resource before starting the migration procedure.
Before the actual migration starts, the user is required to review the to-be-installed AstarteDefaultIngress resource. The actual migration is performed only upon confirmation.`,
	Example: `  astartectl cluster instances migrate replace-voyager --ingress-name <astarte-voyager-ingress-name>`,
	RunE:    replaceVoyagerF,
}

var updateStorageVersionCmd = &cobra.Command{
	Use:   "storage-version",
	Short: "Update Astarte, AVI and Flow CRDs to the v1alpha2 storage version",
	Long: `Update the storage version of Astarte, AstarteVoyagerIngress and Flow CRDs from [v1alpha1, v1alpha2] to v1alpha2.
	
This is NOT a standalone command, please refer to the Astarte documentation on the upgrade to Astarte Operator v22.11 for the complete upgrade procedure.`,
	Example: `  astartectl cluster instances migrate storage-version`,
	RunE:    updateStorageVersionCmdF,
}

var crdsStoredVersionsBeforeUpgrade = []string{"v1alpha1", "v1alpha2"}
var crdsStoredVersionsAfterUpgrade = []string{"v1alpha2"}
var homeDir string
var err error

func init() {
	homeDir, err = os.UserHomeDir()

	if err != nil {
		fmt.Println("Warning: cannot determine the user home directory. Some features may not work properly.", err)
		homeDir = "/tmp"
	}

	InstancesCmd.AddCommand(MigrateCmd)

	replaceVoyagerCmd.PersistentFlags().String("operator-name", "astarte-operator-controller-manager", "The name of the Astarte Operator instance.")
	replaceVoyagerCmd.PersistentFlags().String("ingress-name", "", "The name of the AstarteVoyagerIngress to be migrated. When not set, the first ingress found in the cluster will be selected.")
	replaceVoyagerCmd.PersistentFlags().String("operator-namespace", "kube-system", "The namespace in which the Astarte Operator resides.")
	replaceVoyagerCmd.PersistentFlags().StringP("out", "o", "", "The name of the file in which the AstarteVoyagerIngress custom resource will be saved.")

	MigrateCmd.AddCommand(replaceVoyagerCmd)
	MigrateCmd.AddCommand(updateStorageVersionCmd)

	// API Migration to convert api.astarte-platform.org CRs from one version to another
	MigrateCmd.AddCommand(MigrateApiCmd)
	MigrateApiCmd.AddCommand(tov2alpha1)
	MigrateApiCmd.PersistentFlags().Bool("backup-original-cr", true, "If true, backs up the original CR before migration.")
	MigrateApiCmd.PersistentFlags().String("convert-from-file", "", "If set, reads the Astarte CR from the given file instead of connecting to the cluster.")
	MigrateApiCmd.PersistentFlags().String("namespace", "default", "The namespace in which to look for Astarte instances.")
	MigrateApiCmd.PersistentFlags().Bool("non-interactive", false, "Skip interactive prompts and use placeholder/default values for missing fields.")
}

// migrateApi handles the conversion of all Astarte CR in the cluster (in a namespace) from one API version to another.
func migrateToV2Alpha1(command *cobra.Command, args []string) error {
	sourceVer := astarteV1Alpha3
	destVer := astarteV2Alpha1

	nonInteractive, err := command.Flags().GetBool("non-interactive")
	if err != nil {
		return err
	}

	// Get path p to file from flags
	p, err := command.Flags().GetString("convert-from-file")
	if err != nil {
		return err
	}

	// If a file is provided, read the CR from the file and convert it
	if p != "" {
		fmt.Printf("Converting Astarte CR from file %s\n", p)

		f, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("error opening input file %q: %w", p, err)
		}

		ds := map[string]interface{}{}
		if err := yaml.Unmarshal(f, &ds); err != nil {
			return fmt.Errorf("error parsing input file %q: %w", p, err)
		}

		dsu := &unstructured.Unstructured{Object: ds}

		// Convert the CR
		converted, err := v1alpha3tov2alpha1.V1alpha3toV2alpha1(dsu, nonInteractive)
		if err != nil {
			return fmt.Errorf("failed to convert Astarte CR (%s): %w", dsu.GetName(), err)
		}

		// Save the migrated CR to a file
		d := filepath.Join(homeDir, "astartectl", "cr-migration", destVer.Version, "cr-from-file")
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("failed to create a directory for migrated CRs at %q: %w", d, err)
		}
		convertedFile := filepath.Join(d, fmt.Sprintf("astarte-%s.yaml", destVer.Version))
		if err := migrationutils.DumpResourceToYAMLFile(converted, convertedFile); err != nil {
			return fmt.Errorf("failed to save migrated CR to %s: %w", convertedFile, err)
		}
		fmt.Printf("Migrated CR saved to %s\n", convertedFile)

		return nil
	}

	// Otherwise, connect to the cluster and migrate all instances found
	fmt.Printf("Astarte instances CR will converted from %s to %s\n", sourceVer.Version, destVer.Version)

	backupOriginalCR, err := command.Flags().GetBool("backup-original-cr")
	if err != nil {
		return err
	}

	astarteNamespace, err := command.Flags().GetString("namespace")
	if err != nil {
		return err
	}

	fmt.Printf("Astarte namespace is set to %s\n", astarteNamespace)

	// Get the Astarte instances in the cluster specifically of version v1alpha2
	// If there are no instances of version v1alpha2, then there is nothing to do here.
	astarteList, err := kubernetesDynamicClient.Resource(sourceVer).Namespace(astarteNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(astarteList.Items) == 0 {
		return fmt.Errorf("No Astarte instances of version %s found in the cluster. Nothing to do here.", sourceVer.Version)
	}

	// List the instances found
	fmt.Printf("Found %d Astarte instance(s) in the cluster:\n", len(astarteList.Items))
	for _, astarteObj := range astarteList.Items {
		fmt.Printf("- %s\n", astarteObj.GetName())
	}

	// Warn the user of what is about to happen
	fmt.Printf("You are about to convert all Astarte instances in the namespace %s from %s to %s.\n", astarteNamespace, sourceVer.Version, destVer.Version)
	if nonInteractive {
		fmt.Println("Non-interactive mode enabled. Proceeding without confirmation.")
	} else {
		proceed, err := utils.AskForConfirmation("Are you sure?")
		if err != nil {
			return err
		}
		if !proceed {
			fmt.Println("Ok, nothing left to do here.")
			return nil
		}
	}

	if !backupOriginalCR {
		fmt.Println("Backing up original CRs is disabled. Proceeding without backup.")
	}

	// Convert each CR
	for _, astarteObj := range astarteList.Items {
		fmt.Printf("Converting Astarte CR %s...\n", astarteObj.GetName())

		// Backup original CR if required
		if backupOriginalCR {
			backupDir := filepath.Join(homeDir, "astartectl", "cr-backups", astarteObj.GetNamespace())
			if err := os.MkdirAll(backupDir, 0o755); err != nil {
				return fmt.Errorf("failed to create a backup directory for original CRs at %q: %w", backupDir, err)
			}
			backupFile := filepath.Join(backupDir, fmt.Sprintf("%s-backup-%s.yaml", astarteObj.GetName(), astarteObj.GetResourceVersion()))
			if err := migrationutils.DumpResourceToYAMLFile(&astarteObj, backupFile); err != nil {
				return fmt.Errorf("failed to back up original CR to %s: %w", backupFile, err)
			}
			fmt.Printf("Original CR backed up to %s\n", backupFile)
		}

		// Convert the CR
		converted, err := v1alpha3tov2alpha1.V1alpha3toV2alpha1(&astarteObj, nonInteractive)
		if err != nil {
			return fmt.Errorf("failed to convert Astarte CR (%s): %w", astarteObj.GetName(), err)
		}

		// Save the migrated CR to a file
		d := filepath.Join(homeDir, "astartectl", "cr-migration", destVer.Version, astarteObj.GetNamespace())
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("failed to create a directory for migrated CRs at %q: %w", d, err)
		}
		convertedFile := filepath.Join(d, fmt.Sprintf("%s-%s.yaml", astarteObj.GetName(), destVer.Version))
		if err := migrationutils.DumpResourceToYAMLFile(converted, convertedFile); err != nil {
			return fmt.Errorf("failed to save migrated CR to %s: %w", convertedFile, err)
		}
		fmt.Printf("Converted CR saved to %s\n", convertedFile)
	}

	return nil
}

func updateStorageVersionCmdF(command *cobra.Command, args []string) error {
	crds := []string{"astartes.api.astarte-platform.org", "astartevoyageringresses.api.astarte-platform.org", "flows.api.astarte-platform.org"}

	crdsWithStoredVersions := map[string][]string{}
	//check that required CRDs exist and save their current storedVersions, in case you need to restore them.
	for _, v := range crds {
		if crd, err := kubernetesAPIExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.Background(),
			v, metav1.GetOptions{}); err != nil {
			return err
		} else {
			crdsWithStoredVersions[v] = crd.Status.StoredVersions
		}
	}

	// check that all 3 CRDs have the right storedVersions
	for k, v := range crdsWithStoredVersions {
		if !checkStoredVersionsMatch(v, crdsStoredVersionsBeforeUpgrade) {
			return fmt.Errorf("CRD %s status not consistent with API migration. Refer to the Astarte documentation on the upgrade to Astarte Operator 22.11", k)
		}
	}

	// if all preconditions are met, we can start upgrading the CRDs
	if err := updateAllStoredVersions(crdsWithStoredVersions); err != nil {
		// something went wrong, revert revert revert
		return restoreAllStoredVersions(crdsWithStoredVersions)
	}

	fmt.Println("All CRDs were upgraded successfully!")

	return nil
}

func replaceVoyagerF(command *cobra.Command, args []string) error {
	astarteNamespace, err := command.Flags().GetString("namespace")
	if err != nil {
		return err
	}
	operatorName, err := command.Flags().GetString("operator-name")
	if err != nil {
		return err
	}
	operatorNamespace, err := command.Flags().GetString("operator-namespace")
	if err != nil {
		return err
	}

	// is the migration allowed?
	if err := ensureOperatorMinimumVersionRequirement(operatorName, operatorNamespace); err != nil {
		return err
	}

	// ensure the Voyager CRD is present
	_, err = kubernetesAPIExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.Background(),
		"ingresses.voyager.appscode.com", metav1.GetOptions{})
	if err != nil {
		return err
	}

	aviList, err := kubernetesDynamicClient.Resource(aviV1Alpha1).Namespace(astarteNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if len(aviList.Items) == 0 {
		return errors.New("No AstarteVoyagerIngress resources are present in the cluster. Nothing to do here.")
	}

	// get the first available avi
	aviObject := &aviList.Items[0]

	// and, in case, if the ingress-name is set, override the object
	aviResourceName, err := command.Flags().GetString("ingress-name")
	if err != nil {
		fmt.Println("Error parsing the ingress-name flag. The option will be neglected.", err)
	}
	if aviResourceName != "" {
		for _, item := range aviList.Items {
			if item.GetName() == aviResourceName {
				aviObject = &item
				break
			}
		}
		if aviObject.GetName() != aviResourceName {
			fmt.Printf("Couldn't find the %s resource. Falling back to the %s resource.\n", aviResourceName, aviObject.GetName())
		}
	}

	fmt.Printf("You are about to migrate the AstarteVoyagerIngress named: %s. ", aviObject.GetName())
	shouldMigrate, err := utils.AskForConfirmation("Are you sure?")
	if err != nil {
		return err
	}
	if !shouldMigrate {
		fmt.Println("Ok, nothing left to do here.")
		os.Exit(0)
	}

	// if required, dump the avi custom resource
	if fn := command.Flag("out").Value.String(); fn != "" {
		if err := dumpResourceToYAMLFile(aviObject, fn); err != nil {
			fmt.Println("Warning: failed to save the AstarteVoyagerIngress resource to file: ", err)
		} else {
			fmt.Printf("%s resource successfully written to %s\n", aviObject.GetName(), fn)
		}
	}

	adiName, err := utils.PromptChoice("Choose the new AstarteDefaultIngress name:", "adi", false, false)
	if err != nil {
		return err
	}

	// We are not checking if the secrets are present in the cluster: if they are not, the validation webhook will return an error
	apiSecretName, err := utils.PromptChoice("Insert the name of the secret containing the TLS certificates and keys to connect to the Astarte API and Dashboard:", "", false, false)
	if err != nil {
		return nil
	}
	brokerSecretName, err := utils.PromptChoice("Insert the name of the secret containing the TLS certificates and keys to connect to the Astarte Broker:", "", false, false)
	if err != nil {
		return nil
	}

	ingressClass, err := utils.PromptChoice("Which ingress class should the AstarteDefaultIngress employ?", "nginx", false, false)
	if err != nil {
		return err
	}

	if err := migrateAVIToADI(aviObject, adiName, ingressClass, apiSecretName, brokerSecretName); err != nil {
		return err
	}

	return nil
}

func ensureOperatorMinimumVersionRequirement(operatorName, operatorNamespace string) error {
	operator, err := getAstarteOperator(operatorName, operatorNamespace)
	if err != nil {
		return err
	}

	stringVersion := strings.Split(operator.Spec.Template.Spec.Containers[0].Image, ":")[1]

	if isUnstableVersion(stringVersion) {
		errString := "You are running a snapshot version of the Astarte operator. We cannot support you in this managed migration path. Upgrade to a stable version before trying again."
		return errors.New(errString)
	}

	version, err := semver.NewVersion(stringVersion)
	if err != nil {
		return err
	}

	constraint, err := semver.NewConstraint("> 1.0.0")
	if err != nil {
		return err
	}

	if !constraint.Check(version) {
		return errors.New("Migration is not allowed. Upgrade your Astarte operator to a version > v1.0.0.")
	}

	return nil
}

func handleTLSTerminationAtVernemq(astarteObj *unstructured.Unstructured, secretName string) error {
	return updateAstarteVernemqTLSConfig(astarteObj, secretName, true)
}

func restoreAstarteVernemqTLSConfig(astarteObj *unstructured.Unstructured) error {
	return updateAstarteVernemqTLSConfig(astarteObj, "", false)
}

func updateAstarteVernemqTLSConfig(astarteObj *unstructured.Unstructured, secretName string, sslListener bool) error {
	astarteNamespace := astarteObj.GetNamespace()
	vmqSpecObject, found, err := unstructured.NestedMap(astarteObj.Object["spec"].(map[string]interface{}), "vernemq")
	if err != nil {
		return err
	}
	if !found {
		return errors.New("This is weird! Field \"vernemq\" not found in astarte object.")
	}

	if err := unstructured.SetNestedField(vmqSpecObject, sslListener, "sslListener"); err != nil {
		return err
	}
	if err := unstructured.SetNestedField(vmqSpecObject, secretName, "sslListenerCertSecretName"); err != nil {
		return err
	}
	if err := unstructured.SetNestedMap(astarteObj.Object, vmqSpecObject, "spec", "vernemq"); err != nil {
		return err
	}

	if _, err := kubernetesDynamicClient.Resource(astarteV1Alpha1).Namespace(astarteNamespace).Update(context.Background(), astarteObj, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func migrateAVIToADI(aviObj *unstructured.Unstructured, adiName, ingressClass, apiSecretName, brokerSecretName string) error {
	adiObj := &unstructured.Unstructured{}

	adiObj.SetName(adiName)
	adiObj.SetNamespace(aviObj.GetNamespace())

	adiGVK := schema.GroupVersionKind{
		Group:   "ingress.astarte-platform.org",
		Version: "v1alpha1",
		Kind:    "AstarteDefaultIngress",
	}
	adiObj.SetGroupVersionKind(adiGVK)

	astarteName, _, err := unstructured.NestedString(aviObj.Object, "spec", "astarte")
	if err != nil {
		return err
	}
	if err := unstructured.SetNestedField(adiObj.Object, astarteName, "spec", "astarte"); err != nil {
		return err
	}

	if err := unstructured.SetNestedField(adiObj.Object, ingressClass, "spec", "ingressClass"); err != nil {
		return err
	}

	// prepare specs
	adiApiSpec, err := prepareADIAPISpec(aviObj, apiSecretName)
	if err != nil {
		return err
	}
	adiDashboardSpec, err := prepareADIDashboardSpec(aviObj, apiSecretName)
	if err != nil {
		return err
	}
	adiBrokerSpec, err := prepareADIBrokerSpec(aviObj)
	if err != nil {
		return err
	}

	// notify that, if present, additional annotations are not handled
	maybeNotifyForUnmanagedAnnotations(aviObj)

	// set specs
	if err := unstructured.SetNestedMap(adiObj.Object, adiApiSpec, "spec", "api"); err != nil {
		return err
	}
	if err := unstructured.SetNestedMap(adiObj.Object, adiDashboardSpec, "spec", "dashboard"); err != nil {
		return err
	}
	if err := unstructured.SetNestedMap(adiObj.Object, adiBrokerSpec, "spec", "broker"); err != nil {
		return err
	}

	// check settings before proceeding
	if err := reviewADIAndConfirmMigration(adiObj); err != nil {
		return err
	}

	// perform the migration tasks
	fmt.Println("Starting migration procedure...")
	if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		astarteObj, err := getAstarteInstance(astarteName, adiObj.GetNamespace())
		if err != nil {
			return err
		}
		return handleTLSTerminationAtVernemq(astarteObj, brokerSecretName)
	}); err != nil {
		return err
	}

	// Do not retry on conflicts. We are not expecting conflicts here as this is the first creation of an ADI object.
	// If an error occurs, then restore the astarte vernemq configuration
	if err = createADI(adiObj); err != nil {
		return retry.RetryOnConflict(retry.DefaultRetry, func() error {
			astarteObj, err := getAstarteInstance(astarteName, adiObj.GetNamespace())
			if err != nil {
				return err
			}
			if err = restoreAstarteVernemqTLSConfig(astarteObj); err != nil {
				fmt.Println("Failed to restore the VerneMQ TLS configuration.")
				return err
			}
			fmt.Println("VerneMQ TLS configuration successfully reverted.")
			return nil
		})
	}
	fmt.Println("Migration procedure concluded successfully!")

	if err = deleteAVI(aviObj); err != nil {
		return err
	}

	return nil
}

func reviewADIAndConfirmMigration(adiObj *unstructured.Unstructured) error {
	y, _ := unstructuredToYAML(adiObj)

	fmt.Println("")
	fmt.Println("The following custom resource will be installed. Review it before proceeding.")
	fmt.Println(string(y))

	proceed, err := utils.AskForConfirmation("Do you want to proceed with the migration?")
	if err != nil {
		return err
	}
	if !proceed {
		fmt.Println("Aborting the migration procedure. Your Astarte instance has NOT been modified.")
		os.Exit(0)
	}
	return nil
}

func createADI(adiObj *unstructured.Unstructured) error {
	astarteNamespace := adiObj.GetNamespace()
	if _, err := kubernetesDynamicClient.Resource(adiV1Alpha1).Namespace(astarteNamespace).Create(context.Background(), adiObj, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func deleteAVI(aviObj *unstructured.Unstructured) error {
	fmt.Println("Deleting the old AstarteVoyagerIngress resource...")
	if err := kubernetesDynamicClient.Resource(aviV1Alpha1).Namespace(aviObj.GetNamespace()).Delete(context.Background(), aviObj.GetName(), metav1.DeleteOptions{}); err != nil {
		return err
	}
	fmt.Println("Done!")
	return nil
}

func prepareADIAPISpec(aviObj *unstructured.Unstructured, apiSecretName string) (map[string]interface{}, error) {
	ret := make(map[string]interface{})

	// deploy
	deploy, found, err := unstructured.NestedBool(aviObj.Object, "spec", "api", "deploy")
	if err != nil {
		return ret, err
	}
	if !found {
		// handle the default value
		deploy = true
	}
	if err = unstructured.SetNestedField(ret, deploy, "deploy"); err != nil {
		return ret, err
	}
	if !deploy {
		return ret, nil
	}

	// only load balancer is supported
	serviceType, found, err := unstructured.NestedString(aviObj.Object, "spec", "api", "type")
	if err != nil {
		return ret, err
	}
	if !found {
		// not found, it means the default value (LoadBalancer) is used
		serviceType = "LoadBalancer"
	}
	if serviceType != "LoadBalancer" {
		fmt.Println(`WARNING: your AstarteVoyagerIngress API configuration falls outside of the supported configurations.
Ingresses of type NodePort are not supported. A manual intervention is required. Check the Astarte documentation for further details.`)
		os.Exit(1)
	}

	if lbIp, found, _ := unstructured.NestedString(aviObj.Object, "spec", "api", "loadBalancerIp"); found || lbIp != "" {
		fmt.Println("WARNING: The API ingress load balancer IP will be neglected. To set your API ingress IP set the NGINX ingress controller IP.")
	}

	if err = unstructured.SetNestedField(ret, apiSecretName, "tlsSecret"); err != nil {
		return ret, err
	}

	// Cors
	cors, found, err := unstructured.NestedBool(aviObj.Object, "spec", "api", "cors")
	if err != nil {
		return ret, err
	}
	if !found {
		// handle default
		cors = false
	}
	if err = unstructured.SetNestedField(ret, cors, "cors"); err != nil {
		return ret, err
	}

	// ExposeHousekeeping
	exposeHousekeeping, found, err := unstructured.NestedBool(aviObj.Object, "spec", "api", "exposeHousekeeping")
	if err != nil {
		return ret, err
	}
	if !found {
		// handle default
		exposeHousekeeping = true
	}
	if err = unstructured.SetNestedField(ret, exposeHousekeeping, "exposeHousekeeping"); err != nil {
		return ret, err
	}

	// ServeMetrics
	serveMetrics, found, err := unstructured.NestedBool(aviObj.Object, "spec", "api", "serveMetrics")
	if err != nil {
		return ret, err
	}
	if !found {
		// handle default
		serveMetrics = false
	}
	if err = unstructured.SetNestedField(ret, serveMetrics, "serveMetrics"); err != nil {
		return ret, err
	}

	// ServeMetricsToSubnet
	if serveMetrics {
		serveMetricsToSubnet, found, err := unstructured.NestedString(aviObj.Object, "spec", "api", "serveMetricsToSubnet")
		if err != nil {
			return ret, err
		}
		if !found {
			// handle default
			serveMetricsToSubnet = ""
		}
		if err = unstructured.SetNestedField(ret, serveMetricsToSubnet, "serveMetricsToSubnet"); err != nil {
			return ret, err
		}
	}

	return ret, nil
}

func prepareADIDashboardSpec(aviObj *unstructured.Unstructured, apiSecretName string) (map[string]interface{}, error) {
	ret := make(map[string]interface{})

	// deploy
	deploy, found, err := unstructured.NestedBool(aviObj.Object, "spec", "dashboard", "deploy")
	if err != nil {
		return ret, err
	}
	if !found {
		// handle default
		deploy = true
	}
	if err = unstructured.SetNestedField(ret, deploy, "deploy"); err != nil {
		return ret, err
	}

	// the dashboard is not served. We can just return
	if !deploy {
		return ret, nil
	}

	// ssl
	ssl, found, err := unstructured.NestedBool(aviObj.Object, "spec", "dashboard", "ssl")
	if err != nil {
		return ret, err
	}
	if !found {
		// handle default
		ssl = true
	}
	if err = unstructured.SetNestedField(ret, ssl, "ssl"); err != nil {
		return ret, err
	}

	// tlsSecret
	if ssl {
		if err = unstructured.SetNestedField(ret, apiSecretName, "tlsSecret"); err != nil {
			return ret, err
		}
	}

	// host
	host, found, err := unstructured.NestedString(aviObj.Object, "spec", "dashboard", "host")
	if err != nil {
		return ret, err
	}
	if !found {
		return ret, errors.New("Dashboard host is empty. Something is wrong. Ensure to review your AstarteVoyagerIngress configuration.")
	}
	if err = unstructured.SetNestedField(ret, host, "host"); err != nil {
		return ret, err
	}

	return ret, nil
}

func prepareADIBrokerSpec(aviObj *unstructured.Unstructured) (map[string]interface{}, error) {
	ret := make(map[string]interface{})

	// deploy
	deploy, found, err := unstructured.NestedBool(aviObj.Object, "spec", "broker", "deploy")
	if err != nil {
		return ret, err
	}
	if !found {
		// handle default
		deploy = true
	}
	if err = unstructured.SetNestedField(ret, deploy, "deploy"); err != nil {
		return ret, err
	}

	if !deploy {
		return ret, nil
	}

	// only load balancer is supported
	serviceType, found, err := unstructured.NestedString(aviObj.Object, "spec", "broker", "type")
	if err != nil {
		return ret, err
	}
	if !found {
		// not found, it means the default value (LoadBalancer) is used
		serviceType = "LoadBalancer"
	}
	if serviceType != "LoadBalancer" {
		fmt.Println(`WARNING: your AstarteVoyagerIngress broker configuration falls outside of the supported configurations.
Ingresses of type NodePort are not supported. A manual intervention is required. Check the Astarte documentation for further details.`)
		os.Exit(1)
	}

	if err = unstructured.SetNestedField(ret, serviceType, "serviceType"); err != nil {
		return ret, err
	}

	// loadBalancerIp
	lbIp, found, err := unstructured.NestedString(aviObj.Object, "spec", "broker", "loadBalancerIp")
	if err != nil {
		return ret, err
	}
	if found {
		if err = unstructured.SetNestedField(ret, lbIp, "loadBalancerIP"); err != nil {
			return ret, err
		}
	}

	return ret, nil
}

func maybeNotifyForUnmanagedAnnotations(aviObj *unstructured.Unstructured) {
	apiAs, apiFound, _ := unstructured.NestedMap(aviObj.Object, "spec", "api", "annotationsService")
	brokerAs, brokerFound, _ := unstructured.NestedMap(aviObj.Object, "spec", "broker", "annotationsService")

	if apiFound || brokerFound {
		fmt.Printf("\n-- ATTENTION REQUIRED --\nThe following AstarteVoyagerIngress annotations have been found:\n")
		if len(apiAs) > 0 {
			apiAnnotations, _ := json.MarshalIndent(apiAs, "", "    ")
			fmt.Printf("api\n%s\n", apiAnnotations)
		}
		if len(brokerAs) > 0 {
			brokerAnnotations, _ := json.MarshalIndent(brokerAs, "", "    ")
			fmt.Printf("broker\n%s\n", brokerAnnotations)
		}
		fmt.Println("However, the migration of these annotations is not supported as they are Voyager specific and, as such, they will be dropped.")

	}
}

func updateAllStoredVersions(crds map[string][]string) error {
	for crd := range crds {
		if err := updateStoredVersionsTo(crd, crdsStoredVersionsAfterUpgrade); err != nil {
			fmt.Printf("Error updating %s: %s. Will restore CRDs.", crd, err.Error())
			return err
		}
	}
	return nil
}

func restoreAllStoredVersions(crds map[string][]string) error {
	fmt.Println("Restoring CRDs to the previous state...")
	for crd, oldStoredVersions := range crds {
		if err := updateStoredVersionsTo(crd, oldStoredVersions); err != nil {
			return err
		}
	}
	fmt.Println("Done. Please, check again the upgrade guide for Astarte Operator v22.11 before retrying.")
	return nil
}

func updateStoredVersionsTo(crdName string, newStoredVersions []string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		crd, err := kubernetesAPIExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.Background(),
			crdName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		// if it has the desired storedVersions already, do nothing
		if checkStoredVersionsMatch(crd.Status.StoredVersions, newStoredVersions) {
			return nil
		}

		crd.Status.StoredVersions = newStoredVersions

		if _, err = kubernetesAPIExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().UpdateStatus(context.Background(), crd, metav1.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	})
}

func checkStoredVersionsMatch(currentStoredVersions []string, requiredStoredVersions []string) bool {
	return cmp.Equal(currentStoredVersions, requiredStoredVersions, cmpopts.SortSlices(func(a, b string) bool {
		return a < b
	}))
}
