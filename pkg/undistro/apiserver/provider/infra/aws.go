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
package infra

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
	InstanceType      string `json:"instance_type"`
	AvailabilityZones string `json:"availability_zones"`
}

var (
	regions = []string{
		"us-east-2",
		"us-east-1",
		"us-west-1",
		"us-west-2",
		"af-south-1",
		"ap-east-1",
		"ap-south-1",
		"ap-northeast-3",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
		"ca-central-1",
		"cn-north-1",
		"cn-northwest-1",
		"eu-central-1",
		"eu-west-1",
		"eu-west-2",
		"eu-south-1",
		"eu-west-3",
		"eu-north-1",
		"me-south-1",
		"sa-east-1",
		"us-gov-east-1",
		"us-gov-west-1",
	}
	flavors = map[string]string{
		undistrov1alpha1.EC2.String(): "1.20",
		undistrov1alpha1.EKS.String(): "1.19",
	}

	//go:embed instancetypesaws.json
	machineTypesEmb []byte
)

type metadata struct {
	MachineTypes     []ec2InstanceType `json:"machine_types"`
	ProviderRegions  []string          `json:"provider_regions"`
	SupportedFlavors map[string]string `json:"supported_flavors"`
}

var (
	ErrInvalidProvider  = errors.New("invalid provider, maybe unsupported")
	errGetCredentials   = errors.New("cannot retrieve credentials from secrets")
	errLoadConfig       = errors.New("unable to load SDK config")
	errDescribeKeyPairs = errors.New("error to describe key pairs")
)

func DescribeMachineTypes() (mt []ec2InstanceType, err error) {
	err = json.Unmarshal(machineTypesEmb, &mt)
	return
}

// DescribeSSHKeys retrieve all ssh key names from a region in an account
func DescribeSSHKeys(region string, conf *rest.Config) (res []string, err error) {
	// get credentials from secrets
	k8sClient, err := client.New(conf, client.Options{
		Scheme: scheme.Scheme,
	})
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
		return []string{}, errDescribeKeyPairs
	}

	// filter ssh key names
	for _, kp := range out.KeyPairs {
		res = append(res, *kp.KeyName)
	}
	return res, nil
}

func DescribeMetadata(providerName string) (interface{}, error) {
	var result interface{}

	switch providerName {
	case undistrov1alpha1.Amazon.String():
		pm := metadata{
			ProviderRegions:  regions,
			SupportedFlavors: flavors,
		}
		return pm, nil
	}
	return result, ErrInvalidProvider
}
