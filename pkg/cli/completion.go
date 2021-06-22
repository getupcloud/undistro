/*
Copyright 2020-2021 The UnDistro authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cli

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewCmdCompletion(streams genericclioptions.IOStreams) *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate completion script",
		Long: `To load completions:
	
	Bash:
	
	$ source <(undistro completion bash)
	
	# To load completions for each session, execute once:
	Linux:
	  $ undistro completion bash > /etc/bash_completion.d/undistro
	MacOS:
	  $ undistro completion bash > /usr/local/etc/bash_completion.d/undistro
	
	Zsh:
	
	# If shell completion is not already enabled in your environment you will need
	# to enable it.  You can execute the following once:
	
	$ echo "autoload -U compinit; compinit" >> ~/.zshrc
	
	# To load completions for each session, execute once:
	$ undistro completion zsh > "${fpath[1]}/_undistro"
	
	# You will need to start a new shell for this setup to take effect.
	
	Fish:
	
	$ undistro completion fish | source
	
	# To load completions for each session, execute once:
	$ undistro completion fish > ~/.config/fish/completions/undistro.fish
	`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(streams.Out)
			case "zsh":
				cmd.Root().GenZshCompletion(streams.Out)
			case "fish":
				cmd.Root().GenFishCompletion(streams.Out, true)
			}
		},
	}
}
