/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/getupio-undistro/undistro/client"
	"github.com/spf13/cobra"
)

type generateYAMLOptions struct {
	url           string
	listVariables bool
}

var gyOpts = &generateYAMLOptions{}

var generateYamlCmd = &cobra.Command{
	Use:   "yaml",
	Short: "Process yaml using undistro's yaml processor",
	Long: LongDesc(`
		Process yaml using undistro's yaml processor.

		undistro ships with a simple yaml processor that performs variable
		substitution that takes into account of default values.

		Variable values are either sourced from the undistro config file or
		from environment variables`),

	Example: Examples(`
		# Generates a configuration file with variable values using
		a template from a specific URL.
		undistro generate yaml --from https://github.com/foo-org/foo-repository/blob/master/cluster-template.yaml

		# Generates a configuration file with variable values using
		a template stored locally.
		undistro generate yaml --from ~/workspace/cluster-template.yaml

		# Prints list of variables used in the local template
		undistro generate yaml --from ~/workspace/cluster-template.yaml --list-variables

		# Prints list of variables from template passed in via stdin
		cat ~/workspace/cluster-template.yaml | undistro generate yaml --list-variables
`),

	RunE: func(cmd *cobra.Command, args []string) error {
		return generateYAML(os.Stdin, os.Stdout)
	},
}

func init() {
	// flags for the url source
	generateYamlCmd.Flags().StringVar(&gyOpts.url, "from", "-",
		"The URL to read the template from. It defaults to '-' which reads from stdin.")

	// other flags
	generateYamlCmd.Flags().BoolVar(&gyOpts.listVariables, "list-variables", false,
		"Returns the list of variables expected by the template instead of the template yaml")

	generateCmd.AddCommand(generateYamlCmd)
}

func generateYAML(r io.Reader, w io.Writer) error {
	c, err := client.New(cfgFile)
	if err != nil {
		return err
	}
	options := client.ProcessYAMLOptions{
		ListVariablesOnly: gyOpts.listVariables,
	}
	if gyOpts.url != "" {
		if gyOpts.url == "-" {
			options.ReaderSource = &client.ReaderSourceOptions{
				Reader: r,
			}
		} else {
			options.URLSource = &client.URLSourceOptions{
				URL: gyOpts.url,
			}
		}
	}
	printer, err := c.ProcessYAML(options)
	if err != nil {
		return err
	}
	if gyOpts.listVariables {
		if len(printer.Variables()) > 0 {
			fmt.Fprintln(w, "Variables:")
			for _, v := range printer.Variables() {
				fmt.Fprintf(w, "  - %s\n", v)
			}
		} else {
			fmt.Fprintln(w)
		}
		return nil
	}
	out, err := printer.Yaml()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(out))
	return err
}
