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

	"github.com/astarte-platform/astartectl/cmd/housekeeping"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

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
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.astartectl.yaml)")
	rootCmd.PersistentFlags().StringP("astarte-url", "u", "", "Base url for your Astarte deployment (e.g. https://api.astarte.example.com)")
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("astarte-url"))

	rootCmd.AddCommand(housekeeping.HousekeepingCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	envCfgFile := os.Getenv("ASTARTECTL_CONFIG")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else if envCfgFile != "" {
		viper.SetConfigFile(envCfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".astartectl" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".astartectl")
	}

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("astartectl")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else if envCfgFile != "" {
		// If we explicitly provided a config, print a failure message
		fmt.Printf("Cannot use %s for configuration: %s\n", viper.ConfigFileUsed(), err.Error())
	}
}
