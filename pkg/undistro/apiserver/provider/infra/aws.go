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

import "github.com/getupio-undistro/undistro/apis/app/v1alpha1"

type EC2MachineType struct {
	InstanceType string
	AvailabilityZones []string
}

var (
	Regions = []string {
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
	SupportedFlavors = map[string]string {
		v1alpha1.EC2.String(): "1.20",
		v1alpha1.EKS.String(): "1.19",
	}
	MachineTypes = []EC2MachineType{
		{
			InstanceType:      "t1.micro",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1f",
			},
		},
		{
			InstanceType:      "t2.nano",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1e", "us-east-1f",
			},
		},
		{
			InstanceType:      "t2.micro",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1e", "us-east-1f",
			},
		},
		{
			InstanceType:      "t2.small",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1e", "us-east-1f",
			},
		},
		{
			InstanceType:      "t2.medium",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1e", "us-east-1f",
			},
		},
		{
			InstanceType:      "t2.large",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1e", "us-east-1f",
			},
		},
		{
			InstanceType:      "t2.xlarge",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1e", "us-east-1f",
			},
		},
		{
			InstanceType:      "t2.2xlarge",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1e", "us-east-1f",
			},
		},
		{
			InstanceType:      "t3.nano",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1f", "us-east-1f",
			},
		},
		{
			InstanceType:      "t3.micro",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1f",
			},
		},
		{
			InstanceType:      "t3.small",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1f",
			},
		},
		{
			InstanceType:      "t3.medium",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1f",
			},
		},
		{
			InstanceType:      "t3.large",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1f",
			},
		},
		{
			InstanceType:      "t3.xlarge",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1f",
			},
		},
		{
			InstanceType:      "t3.2xlarge",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1f",
			},
		},
		{
			InstanceType:      "a1.medium",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1d",
			},
		},
		{
			InstanceType:      "a1.large",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1d",
			},
		},
		{
			InstanceType:      "a1.xlarge",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1d",
			},
		},
		{
			InstanceType:      "a1.2xlarge",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1d",
			},
		},
		{
			InstanceType:      "a1.4xlarge",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1d",
			},
		},
		{
			InstanceType:      "a1.metal",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1d",
			},
		},
		{
			InstanceType:      "c1.medium",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d",
			},
		},
		{
			InstanceType:      "c1.xlarge",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d",
			},
		},
		{
			InstanceType:      "c3.large",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d",
			},
		},
		{
			InstanceType:      "c3.xlarge",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d",
			},
		},
		{
			InstanceType:      "c3.2xlarge",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d",
			},
		},
		{
			InstanceType:      "a1.metal",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d",
			},
		},
		{
			InstanceType:      "a1.metal",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d",
			},
		},
		{
			InstanceType:      "a1.metal",
			AvailabilityZones: []string {
				"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d",
			},
		},


	}
)
