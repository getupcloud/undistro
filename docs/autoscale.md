# Enabling autoscaling in your cluster

To enable autoscale in your cluster just add `autoscale` section in your cluster spec like below:

```yaml
  autoscale:
    enabled: true
    minSize: 1
    maxSize: 10
```

You can change minimum and maximum size of autoscaler changing `minSize` and `maxSize` respectively.