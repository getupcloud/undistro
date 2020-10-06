/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/getupio-undistro/undistro/cmd/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// Version provides the version information of undistro
type Version struct {
	ClientVersion *version.Info `json:"undistro"`
}

type versionOptions struct {
	output string
}

var vo = &versionOptions{}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print undistro version.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runVersion()
	},
}

func init() {
	versionCmd.Flags().StringVarP(&vo.output, "output", "o", "", "Output format; available options are 'yaml', 'json' and 'short'")

	RootCmd.AddCommand(versionCmd)
}

func runVersion() error {
	clientVersion := version.Get()
	v := Version{
		ClientVersion: &clientVersion,
	}

	switch vo.output {
	case "":
		fmt.Printf("undistro version: %#v\n", v.ClientVersion)
	case "short":
		fmt.Printf("%s\n", v.ClientVersion.GitVersion)
	case "yaml":
		y, err := yaml.Marshal(&v)
		if err != nil {
			return err
		}
		fmt.Print(string(y))
	case "json":
		y, err := json.MarshalIndent(&v, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(y))
	default:
		return errors.Errorf("invalid output format: %s", vo.output)
	}

	return nil
}
