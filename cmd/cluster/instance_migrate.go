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
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	migrationutils "github.com/astarte-platform/astartectl/cmd/cr-migration"
	"github.com/astarte-platform/astartectl/cmd/cr-migration/v1alpha3tov2alpha1"
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

var homeDir string

func init() {
	var err error
	homeDir, err = os.UserHomeDir()

	if err != nil {
		fmt.Println("Warning: cannot determine the user home directory. Some features may not work properly.", err)
		homeDir = os.TempDir()
	}

	InstancesCmd.AddCommand(MigrateCmd)

	// API Migration to convert api.astarte-platform.org CRs from one version to another
	MigrateCmd.AddCommand(MigrateApiCmd)
	MigrateApiCmd.AddCommand(tov2alpha1)
	MigrateApiCmd.PersistentFlags().Bool("backup-original-cr", true, "If true, backs up the original CR before migration.")
	MigrateApiCmd.PersistentFlags().String("convert-from-file", "", "If set, reads the Astarte CR from the given file instead of connecting to the cluster.")
	MigrateApiCmd.PersistentFlags().String("namespace", "default", "The namespace in which to look for Astarte instances.")
	MigrateApiCmd.PersistentFlags().Bool("non-interactive", false, "Skip interactive prompts and use placeholder/default values for missing fields.")
	MigrateApiCmd.PersistentFlags().String("output-dir", filepath.Join(homeDir, "astartectl", "cr-migration"), "Directory to save migrated CRs")
}

// convertAndSaveCR converts an Astarte CR and saves it to a YAML file.
func convertAndSaveCR(obj *unstructured.Unstructured, outputDir, fileName string, nonInteractive bool) error {
	converted, err := v1alpha3tov2alpha1.RunConversion(obj, nonInteractive)
	if err != nil {
		return fmt.Errorf("failed to convert Astarte CR (%s): %w", obj.GetName(), err)
	}

	// Save to file
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", outputDir, err)
	}

	filePath := filepath.Join(outputDir, fileName)
	if err := migrationutils.DumpResourceToYAMLFile(converted, filePath); err != nil {
		return fmt.Errorf("failed to save migrated CR to %s: %w", filePath, err)
	}

	fmt.Printf("Migrated CR saved to %s\n", filePath)
	return nil
}

// migrateToV2Alpha1 handles the conversion of Astarte CR from version v1alpha3 to v2alpha1.
func migrateToV2Alpha1(command *cobra.Command, args []string) error {
	sourceVer := astarteV1Alpha3
	destVer := astarteV2Alpha1
	fmt.Printf("Astarte instances CR will be converted from %s to %s\n", sourceVer.Version, destVer.Version)

	// Check if we are converting from a file
	filePath, err := command.Flags().GetString("convert-from-file")
	if err != nil {
		return fmt.Errorf("could not get 'convert-from-file' flag: %w", err)
	}

	if filePath != "" {
		return migrateFromFile(command, filePath, destVer.Version)
	}

	// Otherwise, connect to the cluster and migrate all instances found
	return migrateFromCluster(command, sourceVer, destVer)
}

func migrateFromFile(command *cobra.Command, filePath, destVersion string) error {
	fmt.Printf("Converting Astarte CR from file %s\n", filePath)

	nonInteractive, err := command.Flags().GetBool("non-interactive")
	if err != nil {
		return fmt.Errorf("could not get 'non-interactive' flag: %w", err)
	}

	outputDirBase, err := command.Flags().GetString("output-dir")
	if err != nil {
		return fmt.Errorf("could not get 'output-dir' flag: %w", err)
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error opening input file %q: %w", filePath, err)
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(fileContent, &data); err != nil {
		return fmt.Errorf("error parsing input file %q: %w", filePath, err)
	}

	unstructuredData := &unstructured.Unstructured{Object: data}

	name := unstructuredData.GetName()
	if name == "" {
		name = "unnamed-astarte"
	}

	outputDir := filepath.Join(outputDirBase, destVersion, "cr-from-file")
	convertedFile := fmt.Sprintf("%s-%s.yaml", name, destVersion)
	return convertAndSaveCR(unstructuredData, outputDir, convertedFile, nonInteractive)
}

func migrateFromCluster(command *cobra.Command, sourceVer, destVer schema.GroupVersionResource) error {
	backupOriginalCR, err := command.Flags().GetBool("backup-original-cr")
	if err != nil {
		return fmt.Errorf("could not get 'backup-original-cr' flag: %w", err)
	}

	if !backupOriginalCR {
		fmt.Println("Backing up original CRs is disabled. Proceeding without backup.")
	}

	astarteNamespace, err := command.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("could not get 'namespace' flag: %w", err)
	}

	fmt.Printf("Astarte namespace is set to %s\n", astarteNamespace)

	// Get the Astarte instances in the cluster specifically of version v1alpha3
	// If there are no instances of version v1alpha3, then there is nothing to do here.
	astarteList, err := kubernetesDynamicClient.Resource(sourceVer).Namespace(astarteNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list Astarte instances: %w", err)
	}

	if len(astarteList.Items) == 0 {
		return fmt.Errorf("no Astarte instances of version %s found in the cluster. Nothing to do here", sourceVer.Version)
	}

	// List the instances found
	fmt.Printf("Found %d Astarte instance(s) in the cluster:\n", len(astarteList.Items))
	for _, astarteObj := range astarteList.Items {
		fmt.Printf("- %s\n", astarteObj.GetName())
	}

	nonInteractive, err := command.Flags().GetBool("non-interactive")
	if err != nil {
		return fmt.Errorf("could not get 'non-interactive' flag: %w", err)
	}

	outputDirBase, err := command.Flags().GetString("output-dir")
	if err != nil {
		return fmt.Errorf("could not get 'output-dir' flag: %w", err)
	}

	// Convert each CR
	for i := range astarteList.Items {
		astarteObj := astarteList.Items[i]
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

		// Convert and save the CR
		outputDir := filepath.Join(outputDirBase, destVer.Version, astarteObj.GetNamespace())
		convertedFile := fmt.Sprintf("%s-%s.yaml", astarteObj.GetName(), destVer.Version)
		if err := convertAndSaveCR(&astarteObj, outputDir, convertedFile, nonInteractive); err != nil {
			return err
		}
	}

	return nil
}
