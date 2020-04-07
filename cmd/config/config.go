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
	"github.com/spf13/cobra"
)

// ConfigCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:               "config",
	Short:             "Manage local astartectl configuration",
	Long:              `Manage local astartectl configuration.`,
	PersistentPreRunE: configPersistentPreRunE,
}

func init() {
	// No additional vars
}

func configPersistentPreRunE(cmd *cobra.Command, args []string) error {
	return nil
}
