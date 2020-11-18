# Consuming Existing AWS Infrastructure

Normally, UnDistro will create infrastructure on AWS when standing up a new workload cluster. However, it is possible to have UnDistro re-use existing AWS infrastructure instead of creating its own infrastructure. Follow the instructions below to configure UnDistro to consume existing AWS infrastructure.

## Prerequisites

In order to have UnDistro consume existing AWS infrastructure, you will need to have already created the following resources:

* A VPC
* One or more private subnets (subnets that do not have a route to an Internet gateway)
* A public subnet in the same Availability Zone (AZ) for each private subnet (this is required for NAT gateways to function properly)
* A NAT gateway for each private subnet, along with associated Elastic IP addresses
* An Internet gateway for all public subnets
* Route table associations that provide connectivity to the Internet through a NAT gateway (for private subnets) or the Internet gateway (for public subnets)

Note that a public subnet (and associated Internet gateway) are required even if the control plane of the workload cluster is set to use an internal load balancer.

You will need the ID of the VPC and subnet IDs that UnDistro should use. This information is available via the AWS Management Console or the AWS CLI.

Note that there is no need to create an Elastic Load Balancer (ELB), security groups, or EC2 instances; UnDistro will take care of these items.

## Tagging AWS Resources

UnDistro itself does tag AWS resources it creates. The `sigs.k8s.io/cluster-api-provider-aws/cluster/<cluster-name>` (where `<cluster-name>` matches the `metadata.name` field of the Cluster object) tag, with a value of `owned`, tells UnDistro that it has ownership of the resource. In this case, UnDistro will modify and manage the lifecycle of the resource.

When consuming existing AWS infrastructure, the UnDistro AWS provider does not require any tags to be present. The absence of the tags on an AWS resource indicates to UnDistro that it should not modify the resource or attempt to manage the lifecycle of the resource.

However, the built-in Kubernetes AWS cloud provider _does_ require certain tags in order to function properly. Specifically, all subnets where Kubernetes nodes reside should have the `kubernetes.io/cluster/<cluster-name>` tag present. Private subnets should also have the `kubernetes.io/role/internal-elb` tag with a value of 1, and public subnets should have the `kubernetes.io/role/elb` tag with a value of 1. These latter two tags help the cloud provider understand which subnets to use when creating load balancers.

## Configuring the Cluster Specification

```yaml
apiVersion: undistro.io/v1alpha1
kind: Cluster
metadata:
  name: undistro-quickstart
  namespace: default
spec:
  kubernetesVersion: v1.19.1
  cniName: calico
  network:
    podsCIDR: ["192.168.0.0/16"]
    vpc:
      id: vpc-045d216e54ab45efd
    subnets:
      -
        id: subnet-0caf0ecba2c28e452
      - 
        id: subnet-0eba6f266a091a0ea
  controlPlaneNode:
    replicas: 3
    machineType: m5.large
  workerNodes:
    - replicas: 3
      machineType: m5.large
  infrastructureProvider:
    name: aws
    sshKey: undistro
    env:
      - name: AWS_ACCESS_KEY_ID
        value: ""
      - name: AWS_SECRET_ACCESS_KEY
        value: ""
```