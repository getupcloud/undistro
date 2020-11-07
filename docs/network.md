# Networh

UnDistro provides an easy way to configure CIDR blocks for pods and services and, is also possible to create a cluster without any communication to the internet.

## Configuring custom CIDR for pods

Add a valid CIDR to `podsCIDR`

```yaml
network:
    podsCIDR: ["192.168.0.0/16"]
```

## Configuring custom CIDR for services

Add a valid CIDR to `servicesCIDR`

```yaml
network:
    servicesCIDR: ["192.168.0.0/16"]
```

## Creating an internal cluster

To create a cluster without any communication to the internet, just set `internalLB` to **true** in controlPlaneNode section.

```yaml
controlPlaneNode:
    internalLB: true
    replicas: 3
    machineType: t3.large
```