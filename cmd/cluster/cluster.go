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
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// Needed for Cloud Provider Authentication
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// ClusterCmd represents the cluster command
var ClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Interact with a remote Astarte Cluster",
	Long: `Interact with a remote Astarte Cluster. This requires the Astarte Cluster to be installed
with Astarte Operator on Kubernetes. Also, this command is capable of managing an Astarte installation on a Cluster
by installing, upgrading and managing Astarte through its Operator.`,
	PersistentPreRunE: clusterPersistentPreRunE,
}

// InstancesCmd represents the instance command
var InstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Interact with an Astarte Instance on a remote Astarte Cluster",
	Long: `Interact with an Astarte Instance on a remote Astarte Cluster. Through this command it is possible to
manage the entire lifecycle of an Astarte instance, including its installation and maintenance.`,
	Aliases: []string{"instance"},
}

// InstanceCmd represents the cluster instance command

// Set here all custom resources for Astarte
var (
	kubernetesClient              *kubernetes.Clientset
	kubernetesAPIExtensionsClient *apiextensions.Clientset
	kubernetesDynamicClient       dynamic.Interface

	astarteGroupResource = schema.GroupResource{
		Group:    "api.astarte-platform.org",
		Resource: "Astarte",
	}
	astarteV1Alpha1 = schema.GroupVersionResource{
		Group:    "api.astarte-platform.org",
		Version:  "v1alpha1",
		Resource: "astartes",
	}
	aviV1Alpha1 = schema.GroupVersionResource{
		Group:    "api.astarte-platform.org",
		Version:  "v1alpha1",
		Resource: "astartevoyageringresses",
	}
	adiV1Alpha1 = schema.GroupVersionResource{
		Group:    "ingress.astarte-platform.org",
		Version:  "v1alpha1",
		Resource: "astartedefaultingresses",
	}
	crdResource = schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1beta1",
		Resource: "customresourcedefinitions",
	}
	astarteOperatorVersions = map[string]schema.GroupVersionResource{
		"v1alpha1": astarteV1Alpha1,
	}

	astarteResourceClients map[string]dynamic.NamespaceableResourceInterface = make(map[string]dynamic.NamespaceableResourceInterface)
)

func init() {
	defaultKubeconfigPath := ""

	if home, err := homedir.Dir(); err == nil {
		defaultKubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	ClusterCmd.PersistentFlags().StringP("kubeconfig", "k", defaultKubeconfigPath,
		"(optional) absolute path to the kubeconfig file")
	viper.BindPFlag("kubeconfig", ClusterCmd.PersistentFlags().Lookup("kubeconfig"))

	// Add flags which are common to all instances commands
	InstancesCmd.PersistentFlags().StringP("namespace", "n", "astarte", "Namespace of the Astarte resource. Defaults to 'astarte'")

	ClusterCmd.AddCommand(InstancesCmd)
}

func clusterPersistentPreRunE(cmd *cobra.Command, args []string) error {
	// Load in this very order
	kubeconfigEnv := os.Getenv("KUBECONFIG")
	kubeconfig, err := cmd.Flags().GetString("kubeconfig")
	if err != nil {
		return err
	}
	if kubeconfigEnv != "" {
		kubeconfig = kubeconfigEnv
	}
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	// create the clientsets
	kubernetesClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	kubernetesAPIExtensionsClient, err = apiextensions.NewForConfig(config)
	if err != nil {
		return err
	}
	kubernetesDynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	for k, v := range astarteOperatorVersions {
		astarteResourceClients[k] = kubernetesDynamicClient.Resource(v)
	}

	return nil
}
