# Helm Release

Helm Release is responsible to manager [Helm Charts](https://helm.sh/) in a declarative way

## Specification

```yaml
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
  dependencies: # It wait all Helm release declared as dependency be successfully installed (optional)
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
```

## Create Helm release

```bash
undistro create -f hr.yaml
```

## Delete Helm release

```bash
undistro delete -f hr.yaml
```

## Check Helm release

```bash
undistro get hr
```