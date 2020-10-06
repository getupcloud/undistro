/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package main

import (
	"github.com/getupio-undistro/undistro/cmd/undistro/cmd"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	cmd.Execute()
}
