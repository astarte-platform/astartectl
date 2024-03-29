// Copyright © 2019-2023 SECO Mind Srl
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
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate shell completions",
}

var completionBashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate bash completions",
	Long: `
To setup your completions, run

astartectl completion bash > ~/bash_completion.d/astartectl

To configure your bash shell to load completions for each session add to your .bashrc

# ~/.bashrc or ~/.profile
source ~/bash_completion.d/astartectl
`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = rootCmd.GenBashCompletion(os.Stdout)
	},
}

var completionZshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Generate zsh completions",
	Long: `
To setup your completions, make sure that ~/.zsh/completion directory is in your zsh fpath, then run

astartectl completion zsh > ~/.zsh/completion/_astartectl
autoload -U compinit && compinit
`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = rootCmd.GenZshCompletion(os.Stdout)
	},
}

var completionFishCmd = &cobra.Command{
	Use:   "fish",
	Short: "Generate fish completions",
	Long: `
To setup your completions, run

astartectl completion fish > ~/.config/fish/completions/astartectl.fish
`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = rootCmd.GenFishCompletion(os.Stdout, true)
	},
}

func init() {
	completionCmd.AddCommand(completionBashCmd)
	completionCmd.AddCommand(completionZshCmd)
	completionCmd.AddCommand(completionFishCmd)

	rootCmd.AddCommand(completionCmd)
}
