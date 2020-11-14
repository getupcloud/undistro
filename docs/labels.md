# Labels, taints and provider tags

## Labels

It is possible to add labels on controlplane (if not managed) and worker nodes. Adding labels section in node spec.

```yaml
labels:
  key1: value1
  key2: value2
```

## Taints

It is possible to add taints on controlplane (if not managed) and worker nodes. Adding taints section in node spec.

```yaml
taints:
  - key: key1
    value: value1
    effect: effect1
  - key: key2
    value: value2
    effect: effect2
```

## Provider tags

It is possible to add tags on controlplane (if not managed) and worker nodes. Adding providerTags section in node spec. These tags will appears in your infrastructure provider if is supported.

```yaml
providerTags:
  key1: value1
  key2: value2
```