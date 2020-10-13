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
```
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
```
kubectl cluster-info
```

### Install UnDistro CLI
The UnDistro CLI tool handles the lifecycle of a UnDistro management cluster.

Download the latest release from releases page https://github.com/getupio-undistro/undistro/releases.

### Initialize the management cluster

Now that we've got clusterctl installed and all the prerequisites in place, let's transform the Kubernetes cluster
into a management cluster by using `undistro init`.

```
undistro init
```
