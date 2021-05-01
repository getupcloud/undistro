/*
Copyright 2020 The UnDistro authors

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
package main

import (
	"context"
	"flag"
	"os"

	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func main() {
	addr := "0.0.0.0:2020"
	cfgFlags := genericclioptions.NewConfigFlags(true)
	pflags := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfgFlags.AddFlags(pflags)
	pflags.StringVarP(&addr, "apiserver", "a", addr, "The address and port of the UnDistro API server")
	err := pflags.Parse(os.Args[1:])
	if err != nil {
		klog.Exit(err)
	}
	f := cmdutil.NewFactory(cfgFlags)
	server := apiserver.NewServer(f, os.Stdin, os.Stdout, os.Stderr)
	err = server.GracefullyStart(context.Background(), addr)
	if err != nil {
		klog.Exit(err)
	}
}
