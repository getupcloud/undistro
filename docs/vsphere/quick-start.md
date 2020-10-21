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

Now that we've got undistro installed and all the prerequisites in place, let's transform the Kubernetes cluster
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

### vSphere Requirements

Your vSphere environment should be configured with a **DHCP service** in the primary VM Network for your workload Kubernetes clusters.
You will also need to configure one resource pool across the hosts onto which the workload clusters will be provisioned. Every host
in the resource pool will need access to shared storage, such as VSAN in order to be able to make use of MachineDeployments and
high-availability control planes.

To use PersistentVolumes (PV), your cluster needs support for Cloud Native Storage (CNS), which is available in **vSphere 6.7 Update 3** and later.
CNS relies on a shared datastore, such as VSAN.

In addition, to use undistro, you should have a SSH public key that will be inserted into the node VMs for
administrative access, and a VM folder configured in vCenter.

#### vCenter Credentials

In order for `undistro` to bootstrap a management cluster on vSphere, it must be able to connect and authenticate to
vCenter. Ensure you have credentials to your vCenter server (user, password and server URL).

#### Uploading the machine images

It is required that machines provisioned by UnDistro have cloudinit, kubeadm and a container runtime pre-installed.

The machine images are retrievable from public URLs. UnDistro currently supports machine images based on Ubuntu 18.04 and
CentOS 7. A list of published machine images is available ovas. For this guide we'll be deploying Kubernetes
v1.19.1 on Ubuntu 18.04 (link to [machine image][default-machine-image]).


### Creating a self hosted kubernetes on Vsphere

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
  network:
    vpc: network1
  controlPlaneNode:
    replicas: 3
    machineType: ubuntu-1804-kube-v1.19.1
    endpointIP: "" # the IP used for k8s API
  workerNodes:
    - replicas: 3
      machineType: ubuntu-1804-kube-v1.19.1
  infrastructureProvider:
    name: vsphere
    env:
      - name: VSPHERE_USERNAME # The username used to access the remote vSphere endpoint
        value: ""
      - name: VSPHERE_PASSWORD # The password used to access the remote vSphere endpoint
        value: ""
      - name: VSPHERE_SERVER # The vCenter server IP or FQDN
        value: ""
      - name: VSPHERE_DATACENTER # The vSphere datacenter to deploy
        value: ""
      - name: VSPHERE_DATASTORE # The vSphere datastore to deploy
        value: ""
      - name: VSPHERE_RESOURCE_POOL # The vSphere resource pool for your VMs
        value: ""
      - name: VSPHERE_SSH_AUTHORIZED_KEY # The public ssh authorized key on all machines
        value: ""
```

replace empty string to values that satisfies your environment

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
  dependencies:
    -
      apiVersion: undistro.io/v1alpha1
      kind: HelmRelease
      name: nginx
  afterApplyObjects:
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
  chart:
    repository: https://kubernetes.github.io/dashboard
    name: kubernetes-dashboard
    version: 2.6.0
  values:
    ingress:
      enabled: true
    serviceAccount:
      name: undistro-quickstart-dash
```

create a file with content above.

```
undistro create cluster -f cluster-vsphere.yaml
```

### Getting cluster kubeconfig

```
undistro get kubeconfig undistro-quickstart -n default
```

### Cleanup

We'll delete all resources created by this cluster on Vsphere


```bash
ubdistro delete cluster -f cluster-vsphere.yaml
```

<!-- References -->
[vm-template]: https://docs.vmware.com/en/VMware-vSphere/6.7/com.vmware.vsphere.vm_admin.doc/GUID-17BEDA21-43F6-41F4-8FB2-E01D275FE9B4.html
[default-machine-image]: https://storage.googleapis.com/capv-images/release/v1.19.1/ubuntu-1804-kube-v1.19.1.ova
[govc]: https://github.com/vmware/govmomi/tree/master/govc