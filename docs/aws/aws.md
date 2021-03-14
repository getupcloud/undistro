# AWS

## Configure

To configure AWS just add credentials in UnDistro configuration file and run install command

**Configuration file**

replace `<key>` to your keys

```yaml
providers:
  -
    name: aws
    configuration:
      accessKeyID: <key>
      secretAccessKey: <key>
      sessionToken: <key> # if you use 2FA
      region: <key> # default region us-east-1
```

**Install command**

```bash
undistro --config undistro-config.yaml install
```

## Flavors supported

- ec2 (vanilla Kubernetes using AWS EC2 VNs)
- eks (AWS Kubernetes offer)

## Connecting to the nodes via SSH

To access one of the nodes (either a control plane node, or a worker node) via the SSH bastion host, use this command if you are using a non-EKS cluster:

```bash
ssh -i ${CLUSTER_SSH_KEY} ubuntu@<NODE_IP> -o "ProxyCommand ssh -W %h:%p -i ${CLUSTER_SSH_KEY} ubuntu@${BASTION_HOST}"
```

And use this command if you are using a EKS based cluster:

```bash
ssh -i ${CLUSTER_SSH_KEY} ec2-user@<NODE_IP> -o "ProxyCommand ssh -W %h:%p -i ${CLUSTER_SSH_KEY} ubuntu@${BASTION_HOST}"
```

Alternately, users can add a configuration stanza to their SSH configuration file (typically found on macOS/Linux systems as $HOME/.ssh/config):

```bash
Host 10.0.*
User ubuntu # for eks based cluster use ec2-user
IdentityFile <CLUSTER_SSH_KEY>
ProxyCommand ssh -W %h:%p ubuntu@<BASTION_HOST>
```

## Consuming existing AWS infrastructure

UnDistro Cluster lifecycle functionality is provided by [Cluster API project](https://cluster-api.sigs.k8s.io/).

Normally, Cluster API will create infrastructure on AWS when standing up a new workload cluster. However, it is possible to have Cluster API re-use existing AWS infrastructure instead of creating its own infrastructure. Follow the instructions below to configure Cluster API to consume existing AWS infrastructure.

### Prerequisites

In order to have Cluster API consume existing AWS infrastructure, you will need to have already created the following resources:

- A VPC
- One or more private subnets (subnets that do not have a route to an Internet gateway)
- A public subnet in the same Availability Zone (AZ) for each private subnet (this is required for NAT gateways to function properly)
- A NAT gateway for each private subnet, along with associated Elastic IP addresses
- An Internet gateway for all public subnets
- Route table associations that provide connectivity to the Internet through a NAT gateway (for private subnets), or the Internet gateway (for public subnets)

Note that a public subnet (and associated Internet gateway) are required even if the control plane of the workload cluster is set to use an internal load balancer.

You will need the ID of the VPC and subnet IDs that Cluster API should use. This information is available via the AWS Management Console or the AWS CLI.

Note that there is no need to create an Elastic Load Balancer (ELB), security groups, or EC2 instances; Cluster API will take care of these items.

### Tagging AWS Resources

Cluster API itself does tag AWS resources it creates. The `sigs.k8s.io/cluster-api-provider-aws/cluster/<cluster-name> (where <cluster-name> `matches the `metadata.name` field of the Cluster object) tag, with a value of `owned`, tells Cluster API that it has ownership of the resource. In this case, Cluster API will modify and manage the lifecycle of the resource.

When consuming existing AWS infrastructure, the Cluster API AWS provider does not require any tags to be present. The absence of the tags on an AWS resource indicates to Cluster API that it should not modify the resource or attempt to manage the lifecycle of the resource.

However, the built-in Kubernetes AWS cloud provider does require certain tags in order to function properly. Specifically, all subnets where Kubernetes nodes reside should have the `kubernetes.io/cluster/<cluster-name>` tag present. Private subnets should also have the `kubernetes.io/role/internal-elb` tag with a value of `1`, and public subnets should have the `kubernetes.io/role/elb` tag with a value of `1`. These latter two tags help the cloud provider understand which subnets to use when creating load balancers.