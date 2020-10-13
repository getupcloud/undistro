# Quick Start

In this tutorial we'll cover the basics of how to use UnDistro to create one or more Kubernetes clusters and install a helm chart into them.

## Installation

### Common Prerequisites

- Install and setup [kubectl] in your local environment
- Install [Kind] and [Docker]

### Install and/or configure a kubernetes cluster

UnDistro requires an existing Kubernetes cluster accessible via kubectl; during the installation process the
Kubernetes cluster will be transformed into a [management cluster] by installing the UnDistro and the Cluster API [provider components], so it
is recommended to keep it separated from any application workload.

It is a common practice to create a temporary, local bootstrap cluster which is then used to provision
a target [management cluster] on the selected [infrastructure provider].

Choose one of the options below:

1. **Existing Management Cluster**

For production use-cases a "real" Kubernetes cluster should be used with appropriate backup and DR policies and procedures in place. The Kubernetes cluster must be at least v1.19.1.

```bash
export KUBECONFIG=<...>
```

2. **Kind**

<aside class="note warning">

<h1>Warning</h1>

[kind] is not designed for production use.

**Minimum [kind] supported version**: v0.7.0

</aside>

[kind] can be used for creating a local Kubernetes cluster for development environments or for
the creation of a temporary [bootstrap cluster] used to provision a target [management cluster] on the selected infrastructure provider.

The installation procedure depends on the version of kind;

### Kind v0.7.X

Create the kind cluster:

```bash
kind create cluster
```
Test to ensure the local kind cluster is ready:

```bash
kubectl cluster-info
```

### Kind v0.8.X

Export the variable **KIND_EXPERIMENTAL_DOCKER_NETWORK=bridge** to let kind run in the default **bridge** network:
```bash
export KIND_EXPERIMENTAL_DOCKER_NETWORK=bridge
```
Create the kind cluster:

```bash
kind create cluster
```
Test to ensure the local kind cluster is ready:

```bash
kubectl cluster-info
```

### Install UnDistro CLI
The UnDistro CLI tool handles the lifecycle of a UnDistro management cluster.

Download the latest release from releases page https://github.com/getupio-undistro/undistro/releases.

### Initialize the management cluster

Now that we've got clusterctl installed and all the prerequisites in place, let's transform the Kubernetes cluster
into a management cluster by using `undistro init`.

```bash
undistro init
```

If you are running a private version you need to create a YAML with a github token that have access to the repository

```yaml
github-token: <your GitHub token>
```
and then

```
undistro --config undistro.yaml init
```

### Adding on AWS account

Follow the AWS documentation https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html

### Creating a self hosted kubernetes on AWS

We will create a cluster with 3 controlplane node and 3 worker nodes

```yaml
apiVersion: undistro.io/v1alpha1
kind: Cluster
metadata:
  name: undistro-quickstart
  namespace: default
spec:
  kubernetesVersion: v1.19.1
  cniName: calico
  controlPlaneNode:
    replicas: 3
    machineType: t3.large
  workerNodes:
    - replicas: 3
      machineType: t3.large
  infrastructureProvider:
    name: aws
    sshKey: <YOUR SSH KEY NAME>
    env:
      - name: AWS_ACCESS_KEY_ID
        value: ""
      - name: AWS_SECRET_ACCESS_KEY
        value: ""
```

### Installing Helm Charts on the cluster

In this tutorial we will install the Nginx ingress controller and the Kubernestes Dashboard

```yaml
---
apiVersion: undistro.io/v1alpha1
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  clusterName: default/undistro-quickstart
  chart:
    repository: https://kubernetes.github.io/ingress-nginx
    name: ingress-nginx
    version: 2.15.0

---
apiVersion: undistro.io/v1alpha1
kind: HelmRelease
metadata:
  name: kubernetes-dashboard
  namespace: default
spec:
  clusterName: default/undistro-quickstart
  chart:
    repository: https://kubernetes.github.io/dashboard
    name: kubernetes-dashboard
    version: 2.0.4
  values:
    rbac: true
    ingress:
      enabled: true
```
create a file with content above.

```
undistro create cluster -f cluster-aws.yaml
```

### Getting cluster kubeconfig

```
undistro get kubeconfig undistro-quickstart -n default
```

### Bastion

UnDistro creates a bastion host by default, but the ingress are disabled. To be able to access the cluster nodes add your CIDR block and enable ingress traffic in the bastion section of cluster YAML. If bastion section doesn't exists create it inside spec.

```yaml
bastion:
  enabled: true
  disableIngressRules: false
  allowedCIDRBlocks:
    - <YOUR CIDR> OR 0.0.0.0/0 TO ACCEPT ALL
```
**friendly remainder** the SSH with name that referenced in YAML is always necessary to access the nodes via SSH
