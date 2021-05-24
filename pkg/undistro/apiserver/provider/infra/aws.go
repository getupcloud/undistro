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
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	undistroaws "github.com/getupio-undistro/undistro/pkg/cloud/aws"
	"github.com/getupio-undistro/undistro/pkg/scheme"
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
		v1alpha1.EC2.String(): "1.20",
		v1alpha1.EKS.String(): "1.19",
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
	errInvalidProvider  = errors.New("invalid provider, maybe unsupported")
	errGetCredentials   = errors.New("cannot retrieve credentials from secrets")
	errLoadConfig       = errors.New("unable to load SDK config")
	errDescribeKeyPairs = errors.New("error to describe key pairs")
)

func describeMachineTypes() (mt []ec2InstanceType, err error) {
	err = json.Unmarshal(machineTypesEmb, &mt)
	return
}

func isValidInfraProvider(name string) bool {
	return name == v1alpha1.Amazon.String()
}

func DescribeSSHKeys(region string, conf *rest.Config) (res []string, err error) {
	k8sClient, err := client.New(conf, client.Options{
		Scheme: scheme.Scheme,
	})

	creds, _, err := undistroaws.Credentials(context.Background(), k8sClient)

	if err != nil {
		return []string{}, errGetCredentials
	}

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

	e := ec2.New(sess)

	params := ec2.DescribeKeyPairsInput{}
	out, err := e.DescribeKeyPairs(&params)

	if err != nil {
		return []string{}, errDescribeKeyPairs
	}

	for _, kp := range out.KeyPairs {
		res = append(res, *kp.KeyName)
	}

	return res, nil
}

func WriteMetadata(providerName string, w http.ResponseWriter) {
	if !isValidInfraProvider(providerName) {
		http.Error(w, errInvalidProvider.Error(), http.StatusBadRequest)
		return
	}

	switch providerName {
	case v1alpha1.Amazon.String():
		mt, err := describeMachineTypes()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		pm := metadata{
			MachineTypes:     mt,
			ProviderRegions:  regions,
			SupportedFlavors: flavors,
		}

		encoder := json.NewEncoder(w)
		err = encoder.Encode(pm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, errInvalidProvider.Error(), http.StatusBadRequest)
	}
}
