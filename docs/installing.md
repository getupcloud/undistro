# Installing UnDistro

UnDistro requires an existing Kubernetes cluster accessible via kubectl; during the installation process the
Kubernetes cluster will be transformed into a [management cluster](./glossary.md#Management-Cluster) by installing the UnDistro [provider components](./glossary.md#Provider-Components), so it
is recommended to keep it separated from any application workload.

It is a common practice to create a temporary, local bootstrap cluster which is then used to provision
a target [management cluster](./glossary.md#Management-Cluster) on the selected [infrastructure provider](./glossary.md#Infrastructure-Provider).

After [prepare the environment](./installing.md#Prepare-environment) choose one of the options below:

- [**Existing Cluster**](./installing.md#Existing-Cluster)
- [**Kind**](./installing.md#Kind)

## Prepare environment

- Install and setup [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) in your local environment
- Install and setup [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) and [Docker](https://www.docker.com/get-started) **(required just for kind installation method)**

## Existing Cluster

For production use-cases a "real" Kubernetes cluster should be used with appropriate backup and DR policies and procedures in place.

```bash
export KUBECONFIG=<...>
```

## Kind

[Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) is not designed for production use.

**Minimum [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) supported version**: v0.9.0

can be used for creating a local Kubernetes cluster for development environments or for
the creation of local bootstrap cluster which is then used to provision
a target [management cluster](./glossary.md#Management-Cluster) on the selected [infrastructure provider](./glossary.md#Infrastructure-Provider).

## Download UnDistro CLI
The UnDistro CLI tool handles the lifecycle of an UnDistro management cluster.

Download the latest version from the releases page https://github.com/getupio-undistro/undistro/releases.

## Create the configuration file

The configuration change according provider we want to install. Know more in [configuration page](./configuration.md)

## Initialize the management cluster

Now that we've got UnDistro CLI installed and all the prerequisites in place, let's transform the Kubernetes cluster
into a management cluster by using `undistro install`.

```bash
undistro --config undistro-config.yaml install   
```

## Upgrade a provider into management cluster

```bash
undistro upgrade <provider name>
```