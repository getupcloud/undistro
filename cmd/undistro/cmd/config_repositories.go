/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/getupio-undistro/undistro/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

const (
	// RepositoriesOutputYaml is an option used to print the repository list in yaml format.
	RepositoriesOutputYaml = "yaml"
	// RepositoriesOutputText is an option used to print the repository list in text format.
	RepositoriesOutputText = "text"
)

var (
	// RepositoriesOutputs is a list of valid repository list outputs.
	RepositoriesOutputs = []string{RepositoriesOutputYaml, RepositoriesOutputText}
)

type configRepositoriesOptions struct {
	output string
}

var cro = &configRepositoriesOptions{}

var configRepositoryCmd = &cobra.Command{
	Use:   "repositories",
	Args:  cobra.NoArgs,
	Short: "Display the list of providers and their repository configurations.",
	Long: LongDesc(`
		Display the list of providers and their repository configurations.

		undistro ships with a list of known providers; if necessary, edit
		$HOME/.undistro/undistro.yaml file to add new provider or to customize existing ones.`),

	Example: Examples(`
		# Displays the list of available providers.
		undistro config repositories
		
		# Print the list of available providers in yaml format.
		undistro config repositories -o yaml`),

	RunE: func(cmd *cobra.Command, args []string) error {
		return runGetRepositories(cfgFile, os.Stdout)
	},
}

func init() {
	configRepositoryCmd.Flags().StringVarP(&cro.output, "output", "o", RepositoriesOutputText,
		fmt.Sprintf("Output format. Valid values: %v.", RepositoriesOutputs))
	configCmd.AddCommand(configRepositoryCmd)
}

func runGetRepositories(cfgFile string, out io.Writer) error {
	if cro.output != RepositoriesOutputText && cro.output != RepositoriesOutputYaml {
		return errors.Errorf("Invalid output format %q. Valid values: %v.", cro.output, RepositoriesOutputs)
	}

	if out == nil {
		return errors.New("unable to print to nil output writer")
	}

	c, err := client.New(cfgFile)
	if err != nil {
		return err
	}

	repositoryList, err := c.GetProvidersConfig()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(out, 10, 4, 3, ' ', 0)

	switch cro.output {
	case RepositoriesOutputText:
		fmt.Fprintln(w, "NAME\tTYPE\tURL\tFILE")
		for _, r := range repositoryList {
			dir, file := filepath.Split(r.URL())
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name(), r.Type(), dir, file)
		}
	case RepositoriesOutputYaml:
		y, err := yaml.Marshal(repositoryList)
		if err != nil {
			return err
		}
		fmt.Fprint(w, string(y))
	}
	w.Flush()
	return nil
}
