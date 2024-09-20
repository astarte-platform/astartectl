// Copyright © 2020 Ispirata Srl
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

package config

import (
	"fmt"
	"os"

	"github.com/astarte-platform/astartectl/config"
	"github.com/spf13/cobra"
)

var currentContextCmd = &cobra.Command{
	Use:     "current-context",
	Short:   "Shows the current astartectl configuration context",
	Long:    `Shows the current astartectl configuration context.`,
	Args:    cobra.ExactArgs(0),
	RunE:    currentContextF,
	Aliases: []string{"get-current-context"},
}

var currentClusterCmd = &cobra.Command{
	Use:     "current-cluster",
	Short:   "Shows the cluster being used by the current astartectl configuration context",
	Long:    `Shows the cluster being used by the current astartectl configuration context.`,
	Args:    cobra.ExactArgs(0),
	RunE:    currentClusterF,
	Aliases: []string{"get-current-cluster"},
}

var setCurrentContextCmd = &cobra.Command{
	Use:     "set-current-context <context>",
	Short:   "Sets the current astartectl configuration context",
	Long:    `Sets the current astartectl configuration context`,
	Args:    cobra.ExactArgs(1),
	RunE:    setCurrentContextF,
	Aliases: []string{"use-context"},
}

func init() {
	ConfigCmd.AddCommand(currentContextCmd)
	ConfigCmd.AddCommand(currentClusterCmd)
	ConfigCmd.AddCommand(setCurrentContextCmd)
}

func currentContextF(command *cobra.Command, args []string) error {
	baseConfig := config.GetBaseConfig(config.GetConfigDir())
	fmt.Println(baseConfig.CurrentContext)
	return nil
}

func currentClusterF(command *cobra.Command, args []string) error {
	baseConfig := config.GetBaseConfig(config.GetConfigDir())
	currentContext, err := config.LoadContextConfiguration(config.GetConfigDir(), baseConfig.CurrentContext)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(currentContext.Cluster)
	return nil
}

func setCurrentContextF(command *cobra.Command, args []string) error {
	if err := updateCurrentContext(args[0]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Context switched to %s\n", args[0])
	return nil
}

func updateCurrentContext(newCurrentContext string) error {
	config.UpdateBaseConfigWithContext(config.GetConfigDir(), newCurrentContext)
	return nil
}
