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

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/astarte-platform/astartectl/cmd/appengine"
	"github.com/astarte-platform/astartectl/cmd/cluster"
	configcmd "github.com/astarte-platform/astartectl/cmd/config"
	"github.com/astarte-platform/astartectl/cmd/housekeeping"
	"github.com/astarte-platform/astartectl/cmd/pairing"
	"github.com/astarte-platform/astartectl/cmd/realm"
	"github.com/astarte-platform/astartectl/cmd/utils"
	"github.com/astarte-platform/astartectl/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgContext string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "astartectl",
	Short: "astartectl",
	Long: `astartectl helps you manage your Astarte deployment.

See below for usage details`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().String("config-dir", "", fmt.Sprintf("config directory (default is %s)", config.GetDefaultConfigDir()))
	rootCmd.PersistentFlags().StringVar(&cfgContext, "context", "", "Configuration context to use. When not specified, defaults to current context.")
	rootCmd.PersistentFlags().StringP("astarte-url", "u", "", "Base url for your Astarte deployment (e.g. https://api.astarte.example.com)")
	rootCmd.PersistentFlags().StringP("token", "t", "", "Token for authenticating against Astarte APIs. When set, it takes precedence over any private key setting. Claims in the token have to match the permissions needed for the individual command.")
	if err := viper.BindPFlag("config-dir", rootCmd.PersistentFlags().Lookup("config-dir")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("astarte-url"))
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))

	rootCmd.AddCommand(housekeeping.HousekeepingCmd)
	rootCmd.AddCommand(pairing.PairingCmd)
	rootCmd.AddCommand(realm.RealmManagementCmd)
	rootCmd.AddCommand(utils.UtilsCmd)
	rootCmd.AddCommand(appengine.AppEngineCmd)
	rootCmd.AddCommand(cluster.ClusterCmd)
	rootCmd.AddCommand(configcmd.ConfigCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// If the config does not exist, do not warn - it's simply not there.
	if err := config.ConfigureViper(cfgContext); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "warn: Error while loading configuration: %s\n", err.Error())
	}

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("astartectl")
	viper.AutomaticEnv() // read in environment variables that match
}
