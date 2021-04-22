# 1 - Quick Start

Siga esses passos e crie seu primeiro cluster com a UnDistro facilmente

# 2 - Glossary

Find below a glossary to help to clarify the doc content

## Management Cluster

The cluster where UnDistro and provider components are installed

## Provider Components

UnDistro and its dependencies

## Infrastructure Provider

UnDistro part that is responsible to communicate with any infrastructure that UnDistro supports

## Core Provider

UnDistro and all required dependencies

# 3 - Introduction

## What is UnDistro (will be in version 1.0.0)?

UnDistro is an enterprise software that automates multicloud, on-prem, and edge operations with a single management UI.

UnDistro automates thousands of Kubernetes clusters across multi-cloud, on-prem and edge with unparalleled resilience. Deploy, manage and run multiple Kubernetes clusters with our platform. On your preferred infrastructure.

UnDistro Kubernetes Platform is directly integrated with leading cloud providers, and runs even in your own datacenter.

By providing managed Kubernetes clusters for your infrastructure, UnDistro makes Kubernetes as easy as it can be. UnDistro empowers you to take advantage of all the advanced features that Kubernetes has to offer and increases the speed, flexibility and scalability of your deployment workflow.

UnDistro provides live updates of your Kubernetes cluster without disrupting your daily business.

# 4 - Installing UnDistro

UnDistro requires an existing Kubernetes cluster accessible via kubectl. During the installation process
the Kubernetes cluster will be transformed into a [management cluster](./docs#Management-Cluster) by installing the UnDistro [provider components](./docs#Provider-Components), so it
is recommended to keep it separated from any application workload.

It is a common practice to create a temporary, local bootstrap cluster which is then used to provision
a target [management cluster](./docs#management-cluster) on the selected [infrastructure provider](./docs#infrastructure-provider).

After [prepare the environment](./docs#prepare-environment) choose one of the options below:

- [**Existing Cluster**](./docs#existing-cluster)
- [**Kind**](./docs#kind)

## Prepare environment

- Install and setup [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) in your local environment
- Install and setup [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) and [Docker](https://www.docker.com/get-started) **(required just for kind installation method)**

## Existing Cluster

For production use-cases a "real" Kubernetes cluster should be used with appropriate backup and DR policies and procedures in place.

~~~bash
export KUBECONFIG={...}
~~~

## Kind

[Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) is not designed for production use.

**Minimum [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) supported version**: v0.9.0

Can be used for creating a local Kubernetes cluster for development environments or for the creation of local bootstrap cluster which is then used to provision
a target
 [management cluster](./docs#Management-Cluster) on the selected [infrastructure provider](./docs#Infrastructure-Provider).

## Download UnDistro CLI

The UnDistro CLI tool handles the lifecycle of an UnDistro management cluster.

Download the latest version from the releases page: https://github.com/getupio-undistro/undistro/releases.

## Create the configuration file

The configuration changes according to provider we want to install. Know more in [configuration page](./docs#configuration)

## Initialize the management cluster

Now that we have got UnDistro CLI installed and all the prerequisites are in place, let's transform the Kubernetes cluster
into a management cluster by using **undistro install**.

~~~bash
undistro --config undistro-config.yaml install   
~~~

## Upgrade a provider into management cluster

~~~bash
undistro upgrade {provider name}
~~~

# 5 - Configuration

Configuration file is used by UnDistro just in the install and move operations.

## Reference

~~~go
type Config struct {
	Credentials   Credentials `mapstructure:"credentials" json:"credentials,omitempty"`
	CoreProviders [ ] Provider  `mapstructure:"coreProviders" json:"coreProviders,omitempty"`
	Providers     [ ] Provider  `mapstructure:"providers" json:"providers,omitempty"`
}
type Credentials struct {
	Username string `mapstructure:"username" json:"username,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty"`
}

type Provider struct {
	Name          string            `mapstructure:"name" json:"name,omitempty"`
	Configuration map[string]string `mapstructure:"configuration" json:"configuration,omitempty"`
}
~~~

### Config

|Name       |Type       |Description|
|-----------|-----------|-----------|
|credentials|Credentials|The registry credentials to use private images|
|coreProviders|[ ] Provider|Core providers can be undistro, cert-manager, cluster-api|
|providers|[ ] Provider| providers can configure any supported infrastructure provider|

### Credentials

|Name       |Type       |Description|
|-----------|-----------|-----------|
|username|string|The registry username|
|password|string|The registry password|

### Provider

|Name       |Type       |Description|
|-----------|-----------|-----------|
|name|string|Provider name|
|configuration|map[string]string|Change according provider name. See provider docs|

# 6 - Providers 

# AWS

## Configure

To configure AWS just add AWS credentials with administrator permissions in UnDistro configuration file and run install command

**Configuration file**

replace **{key}** to your keys

~~~yaml
providers:
  -
    name: aws
    configuration:
      accessKeyID: {key}
      secretAccessKey: {key}
      sessionToken: {key} # if you use 2FA
      region: {key} # default region us-east-1
~~~

**Install command**

~~~bash
undistro --config undistro-config.yaml install
~~~

## Flavors supported

- ec2 (vanilla Kubernetes using AWS EC2 VMs)
- eks (AWS Kubernetes offer)

## Connecting to the nodes via SSH

To access one of the nodes (either a control plane node, or a worker node) via the SSH bastion host, use this command if you are using a non-EKS cluster:

~~~bash
ssh -i ${CLUSTER_SSH_KEY} ubuntu@{NODE_IP} -o "ProxyCommand ssh -W %h:%p -i ${CLUSTER_SSH_KEY} ubuntu@${BASTION_HOST}"
~~~

And use this command if you are using a EKS based cluster:

~~~bash
ssh -i ${CLUSTER_SSH_KEY} ec2-user@{NODE_IP} -o "ProxyCommand ssh -W %h:%p -i ${CLUSTER_SSH_KEY} ubuntu@${BASTION_HOST}"
~~~

Alternately, users can add a configuration stanza to their SSH configuration file (typically found on macOS/Linux systems as $HOME/.ssh/config):

~~~bash
Host 10.0.*
User ubuntu # for eks based cluster use ec2-user
IdentityFile {CLUSTER_SSH_KEY}
ProxyCommand ssh -W %h:%p ubuntu@{BASTION_HOST}
~~~

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

Cluster API itself does tag AWS resources it creates. The **sigs.k8s.io/cluster-api-provider-aws/cluster/{cluster-name}** (where *{cluster-name}* matches the *metadata.name* field of the Cluster object) tag, with a value of **owned**, tells Cluster API that it has ownership of the resource. In this case, Cluster API will modify and manage the lifecycle of the resource.

When consuming existing AWS infrastructure, the Cluster API AWS provider does not require any tags to be present. The absence of the tags on an AWS resource indicates to Cluster API that it should not modify the resource or attempt to manage the lifecycle of the resource.

However, the built-in Kubernetes AWS cloud provider doesn’t  require certain tags in order to function properly. Specifically, all subnets where Kubernetes nodes 
reside should have the **kubernetes.io/cluster/{cluster-name}** tag present. Private subnets should also have the **kubernetes.io/role/internal-elb** tag with a value of **1**, and public subnets should have the **kubernetes.io/role/elb** tag with a value of **1**. These latter two tags help the cloud provider understand which subnets to use when creating load balancers.

# 7 - Cluster

The cluster object is responsible for creating and managing a Kubernetes cluster.

## Specification

~~~yaml
apiVersion: app.undistro.io/v1alpha1
kind: Cluster
metadata:
  name: undistro-quickstart # Cluster name
  namespace: default # Namespace where object is created in management cluster
spec:
  kubernetesVersion: v1.19.5 # Version of kubernetes
  controlPlane: # Control plane specification (it's not used by all infrastructure provider and flavors)
    internalLB: true # Make kubernetes API available just in private network (default false)
    replicas: 1 # Number of machines used as control plane
    machineType: t3.medium # Machine type change according infrastructure provider
    subnet: subnetID # Specify the subnet for control plane machines (optional)
    labels: # Add kubernetes labels in control plane nodes (optional)
      key1: val1
      key2: val2
    providerTags: # Many cloud provider support tags, so you can add here (optional)
      key1: val1
      key2: val2
    taints: # Add kubernetes taints in control plane nodes (optional)
      - key: key1
        value: val1
        effect: NoSchedule
  workers:
    - replicas: 1 # Number of machines used as worker in this node pool
      machineType: t3.medium # Machine type change according infrastructure provider
      subnet: subnetID # Specify the subnet for node pool machines (optional)
      labels: # Add kubernetes labels in node pool nodes (optional)
        key1: val1
        key2: val2
      providerTags: # Many cloud provider support tags, so you can add here (optional)
        key1: val1
        key2: val2
      taints: # Add kubernetes taints in node pool nodes (optional)
        - key: key1
          value: val1
          effect: NoSchedule
      infraNode: true # Enable infra nodes on this node pool nodes (optional)
      autoscaling: # Enable autoscaling (optional)
        enabled: true
        minSize: 1 # Node pool minimum size
        maxSize: 10 # Node pool maximum size
  bastion: # Enable bastion host (enabled by default if SSH key is passed in infrastructureProvider)
    enabled: true
    instanceType: t2.micro
    allowedCIDRBlocks: # Allowed CIDR blocks to access bastion host
      - "0.0.0.0/0" 
  infrastructureProvider:
    name: aws # Required providers supported for now: aws
    sshKey: undistro # Key pair name available on aws
    flavor: ec2 # Required aws flavors supported for now: ec2 or eks
    region: us-east-1 # Required aws available regions
  network: # customize cluster network
    apiServerPort: 6443
    services: [""] # customize CIDR used for services
    pods: [""] # customize CIDR used for pods
    serviceDomain: "svc.cluster.local"
    multiZone: true # Enable cluster in multiple cloud zones
    vpc:
      id: vpcID # Create cluster using already created vpc
      cidrBlock: 10.0.0.0/16 # Customize VPC CIDR block
      zone: s-east-1a # Specify a zone for vpc
    subnets:
      - id: subnetID # Create cluster using already created subnet
        cidrBlock: 10.0.0.0/16 # Customize subnet CIDR block
        zone: s-east-1a # Specify a zone for subnet
        isPublic: false # Specify if subnet is public
~~~

## Create a cluster

~~~bash
undistro create -f cluster.yaml
~~~

## Delete a cluster

~~~bash
undistro delete -f cluster.yaml
~~~

## Consuming existing infrastructure

Check infrastructure provider specific page to see the prerequisites.

## Get cluster kubeconfig

~~~bash
undistro get kubeconfig {cluster name} -n namespace
~~~

## See cluster events

~~~bash
undistro show-progress {cluster name} -n namespace
~~~

## Convert the created cluster into a management cluster

If you are using local cluster as a management cluster you can use move command to convert created cluster into a management cluster

~~~bash
undistro move {cluster name} -n namespace
~~~

## Check cluster

~~~bash
undistro get cl
~~~

## A special thanks

A special thanks for [Cluster API project](https://cluster-api.sigs.k8s.io/) to helps UnDistro to provide the cluster lifecycle functionality.

# 8 - Helm Release

The HelmRelease object is responsible to manage [Helm Charts](https://helm.sh/) in a declarative way

## Specification

~~~yaml
apiVersion: app.undistro.io/v1alpha1
kind: HelmRelease
metadata:
    name: kubernetes-dashboard # Object name
    namespace: default # Object namespace
spec:
  chart:
    secretRef: # Set reference to secret that contains repository credentials if repository is private (optional)
      name: name # Secret name
      namespace: namespace # Secret namespace
    repository: https://kubernetes.github.io/dashboard # Chart repository
    name: kubernetes-dashboard # Chart name
    version: 3.0.2 # Chart version
  clusterName: default/undistro-quickstart # Reference of the cluster where helm chart will be installed in format namespace/name
  autoUpgrade: true # Enable auto upgrade chart. It does not upgrade major versions (optional)
  dependencies: # It waits all Helm release declared as dependency be successfully installed (optional)
    -
      apiVersion: app.undistro.io/v1alpha1
      kind: HelmRelease
      name: nginx
      namespace: default
  afterApplyObjects: # List of Kubernetes to be applied after chart installation (optional)
    -
      apiVersion: rbac.authorization.k8s.io/v1
      kind: ClusterRoleBinding
      metadata:
        name: dashboard-access
      roleRef:
        apiGroup: rbac.authorization.k8s.io
        kind: ClusterRole
        name: cluster-admin
      subjects:
        - kind: ServiceAccount
          name: undistro-quickstart-dash
          namespace: default
  beforeApplyObjects: # List of Kubernetes to be applied before chart installation (optional)
    -
      apiVersion: rbac.authorization.k8s.io/v1
      kind: ClusterRoleBinding
      metadata:
        name: dashboard-access
      roleRef:
        apiGroup: rbac.authorization.k8s.io
        kind: ClusterRole
        name: cluster-admin
      subjects:
        - kind: ServiceAccount
          name: undistro-quickstart-dash
          namespace: default 
  valuesFrom: # Set chart values from a Secret or ConfigMap (optional)
    - name: name # Object name
      kind: Secret # Secret or ConfigMap
      targetPath: key # Secret or ConfigMap key
      valuesKey: key # Chart values file key
      optional: true # Ignore if not found
  values: # Chart values (optional)
    ingress:
      enabled: true
    serviceAccount:
      name: undistro-quickstart-dash
~~~

## Create Helm release

~~~bash
undistro create -f hr.yaml
~~~

## Delete Helm release

~~~bash
undistro delete -f hr.yaml
~~~

## Check Helm release

~~~bash
undistro get hr
~~~

# 9 - Architecture

The overarching architecture of UnDistro is centered around a "management plane". This plane is expected to serve as a single interface upon which administrators can create, scale, upgrade, and delete Kubernetes clusters. At a high level view, the management plane + created clusters should look something like this:

![Image of Architecture](https://github.com/getupio-undistro/undistro/blob/b02153cceba365ed7dbc02ca12ed5a484bb50d12/website/src/assets/images/arch.png?raw=true)

# 10 - Diagrams

## Install

![Image of Install](https://github.com/getupio-undistro/undistro/blob/b02153cceba365ed7dbc02ca12ed5a484bb50d12/website/src/assets/images/install.png?raw=true)

## Usage

![Image of Usage](https://github.com/getupio-undistro/undistro/blob/b02153cceba365ed7dbc02ca12ed5a484bb50d12/website/src/assets/images/usage.png?raw=true)

# 11 - Community

- [Issue tracker](https://github.com/getupio-undistro/undistro/issues)
- [Forum](https://github.com/getupio-undistro/undistro/discussions)
