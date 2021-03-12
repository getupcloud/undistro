# Cluster

Cluster is responsible for create and manage a Kubernetes cluster.

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
```

## Consuming Existing Infrastructure

Check infrastructure provider specific page to see the prerequisites.

## Get cluster kubeconfig

```bash
undistro get kubeconfig <cluster name> -n namespace
```

## See cluster events

```bash
undistro show-progress <cluster name> -n namespace
```

## Convert the created cluster into a management cluster

If you are using local cluster as a management cluster you can use move command to convert created cluster into a management cluster

```bash
undistro move <cluster name> -n namespace
```