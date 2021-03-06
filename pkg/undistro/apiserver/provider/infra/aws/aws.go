/*
Copyright 2021 The UnDistro authors

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
package aws

import (
	"context"
	_ "embed"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	undistrov1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	undistroaws "github.com/getupio-undistro/undistro/pkg/cloud/aws"
	"github.com/getupio-undistro/undistro/pkg/scheme"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ec2InstanceType struct {
	InstanceType      string   `json:"instanceType,omitempty"`
	AvailabilityZones []string `json:"availabilityZones,omitempty"`
	VCPUs             int      `json:"vcpus,omitempty"`
	Memory            float64  `json:"memory,omitempty"`
}

type flavor struct {
	Name               string   `json:"name"`
	KubernetesVersions []string `json:"kubernetesVersion"`
}

var (
	regions = []string{
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-south-1",
		"ap-southeast-1",
		"ap-northeast-2",
		"ca-central-1",
		"eu-central-1",
		"eu-west-1",
		"eu-west-2",
		"eu-west-3",
		"sa-east-1",
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
	}
	flavors = []flavor{
		{
			Name: undistrov1alpha1.EC2.String(),
			KubernetesVersions: []string{
				"v1.18.19", "v1.18.20", "v1.21.2", "v1.19.12", "v1.20.8",
			},
		},
		{
			Name:               undistrov1alpha1.EKS.String(),
			KubernetesVersions: []string{"v1.20.4", "v1.19.8", "v1.18.16"},
		},
	}

	//go:embed instancetypesaws.json
	machineTypesEmb []byte
)

var (
	errGetCredentials   = errors.New("cannot retrieve credentials from secrets")
	errLoadConfig       = errors.New("unable to load SDK config")
	errInvalidPageRange = errors.New("invalid page range")
	errRegionRequired   = errors.New("region is required")
	ErrNoProviderMeta   = errors.New("meta is required. supported are " +
		"['sshKeys', 'regions', 'machineTypes', 'supportedFlavors']")
)

type metaParam string

const (
	SShKeysMeta          = metaParam("sshKeys")
	RegionsMeta          = metaParam("regions")
	MachineTypesMeta     = metaParam("machineTypes")
	SupportedFlavorsMeta = metaParam("supportedFlavors")
)

type PagerResponse struct {
	Page         int
	TotalPages   int
	MachineTypes []ec2InstanceType
}

func DescribeMeta(config *rest.Config, region, meta string, page, itemsPerPage int) (interface{}, error) {
	switch meta {
	case string(RegionsMeta):
		return regions, nil
	case string(SShKeysMeta):
		if region == "" {
			return nil, errRegionRequired
		}
		keys, err := describeSSHKeys(region, config)
		return keys, err
	case string(MachineTypesMeta):
		total, mts, err := describeMachineTypes(page, itemsPerPage)

		pr := PagerResponse{
			Page:         page,
			MachineTypes: mts,
			TotalPages:   total,
		}
		return pr, err
	case string(SupportedFlavorsMeta):
		return flavors, nil
	}
	return nil, ErrNoProviderMeta
}

// describeSSHKeys retrieve all ssh key names from a region in an account
func describeSSHKeys(region string, restConf *rest.Config) (res []string, err error) {
	// get credentials from secrets
	k8sClient, err := client.New(restConf, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return []string{}, errGetCredentials
	}
	creds, _, err := undistroaws.Credentials(context.Background(), k8sClient)
	if err != nil {
		return []string{}, errGetCredentials
	}

	// instantiate config and session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			creds.AccessKeyID,
			creds.SecretAccessKey,
			creds.SessionToken,
		),
	})
	if err != nil {
		return []string{}, errLoadConfig
	}

	// get ssh keys from ec2
	ec2Client := ec2.New(sess)
	params := ec2.DescribeKeyPairsInput{}
	out, err := ec2Client.DescribeKeyPairs(&params)
	if err != nil {
		return []string{}, err
	}

	// filter ssh key names
	for _, kp := range out.KeyPairs {
		res = append(res, *kp.KeyName)
	}
	return res, nil
}

func machineTypes() (mt []ec2InstanceType, err error) {
	err = json.Unmarshal(machineTypesEmb, &mt)
	return
}

// describeMachineTypes receives an integer page value and returns 10 items
func describeMachineTypes(page, itemsPerPage int) (total int, it []ec2InstanceType, err error) {
	// retrieve all machine types
	mt, err := machineTypes()
	if err != nil {
		return
	}

	// calculate total pages
	totalMts := len(mt)
	total = totalMts / itemsPerPage
	remnant := totalMts % itemsPerPage
	if remnant != 0 {
		total += 1
	}

	// pages start at 1, can't be 0 or less.
	start := (page - 1) * itemsPerPage
	stop := start + itemsPerPage
	if start > len(mt) {
		return total, it, errInvalidPageRange
	}
	if stop > len(mt) {
		stop = len(mt)
	}

	return total, mt[start:stop], nil
}
