# Cluster

Cluster is reponsible for create and manage a Kubernetes cluster.

## Specification

```yaml
apiVersion: app.undistro.io/v1alpha1
kind: Cluster
metadata:
  name: undistro-quickstart # Cluster name
  namespace: default # Namespace where object is created in management cluster
spec:
  kubernetesVersion: v1.19.5 # Version of kubernetes
  controlPlane: # Control plane specification (it's not used by all infrastructure provider and flavors)
    replicas: 1 # Number of machines used as control plane
    machineType: t3.medium # Machine type change according infrastructure provider
  workers:
    - replicas: 1 # Number of machines used as worker in this node pool
      machineType: t3.medium # Machine type change according infrastructure provider
  bastion:
    enabled: true
    allowedCIDRBlocks:
      - "0.0.0.0/0"
  infrastructureProvider:
    name: aws
    sshKey: undistro
    flavor: ec2
    region: us-east-1
```