#!/bin/bash

echo ---------------------------
echo Creating Management Cluster
echo ---------------------------

MNG_CLUSTER=mng-cluster
WRK_CLUSTER=myeks
read -ep "Cluster name [$WRK_CLUSTER]: " cluster
WRK_CLUSTER=${cluster:-$WKR_CLUSTER}

export KUBECONFIG=$PWD/kubeconfig.manager

if kind get clusters 2>/dev/null| grep -q "^${MNG_CLUSTER}\$"; then
    read -ep "Kind cluster already exists. Destroy it? [Y/n]: " -n 1 c
    if [ "$c" == y ] || [ "$c" == Y ]; then
        kind delete cluster --name $MNG_CLUSTER
    fi
fi

kind create cluster --name $MNG_CLUSTER

echo -----------------------------------------
echo Installing UnDistro on Management Cluster
echo -----------------------------------------

[ -v AWS_ACCESS_KEY_ID ] || read -ep "AWS_ACCESS_KEY_ID: " AWS_ACCESS_KEY_ID
[ -v AWS_SECRET_ACCESS_KEY ] || read -esp "AWS_SECRET_ACCESS_KEY: " AWS_SECRET_ACCESS_KEY
[ -v AWS_REGION ] || read -ep "AWS_REGION: " AWS_REGION

cat > undistro.yaml <<EOF
providers:
- name: aws
  configuration:
    accessKeyID: ${AWS_ACCESS_KEY_ID}
    secretAccessKey: ${AWS_SECRET_ACCESS_KEY}
    region: ${AWS_REGION}
EOF

undistro --config undistro.yaml install 2>/dev/null
#watch kubectl get pods --all-namespaces

echo ---------------------------
echo Creating EKS Worker Cluster
echo ---------------------------

[ -v AWS_SSH_KEY_NAME ] || read -ep "AWS_SSH_KEY_NAME: " AWS_SSH_KEY_NAME

cat > cluster-eks.yaml <<EOF
apiVersion: app.undistro.io/v1alpha1
kind: Cluster
metadata:
  name: ${WRK_CLUSTER}
  namespace: default
spec:
  kubernetesVersion: v1.18.9
  workers:
    - replicas: 2
      machineType: t3.large
  infrastructureProvider:
    name: aws
    sshKey: ${AWS_SSH_KEY_NAME}
    region: ${AWS_REGION}
    flavor: eks
EOF

undistro create -f cluster-eks.yaml

echo "You can follow events with:"
echo "  $ KUBECONFIG=$PWD/kubeconfig.manager kubectl get events -n default -w"
echo
echo Waiting cluster to finish create...
echo kubectl get cluster -n default $WRK_CLUSTER -w
echo
kubectl get cluster -n default $WRK_CLUSTER -w | while read line; do
    echo "$line"
    if grep -q True <<<"$line"; then
        kill -INT ${PIPESTATUS[0]}
        break
    fi
done

helmrelease-wait-install()
{
    local name=$1
    local ns=$2
    local file=$2
    #kubectl apply -f $file

    echo Waiting HelmRelease $name to install...
    echo kubectl get helmrelease -n default $name -w

    kubectl get helmrelease -n default $name -w | while read line; do
        echo "$line"
        if grep -q True <<<"$line"; then
            kill -INT ${PIPESTATUS[0]}
            return
        fi
    done
}

echo -----------------------------------
echo Installing Helm Chart ingress-traefik
echo -----------------------------------

cat > helm-release-traefik-nginx.yaml <<EOF
apiVersion: app.undistro.io/v1alpha1
kind: HelmRelease
metadata:
    name: ingress-traefik
    namespace: default
spec:
  chart:
    repository: https://kubernetes.github.io/ingress-traefik
    name: ingress-traefik
    version: 3.19.0
  targetNamespace: ingress-traefik
  clusterName: default/$WRK_CLUSTER
  values:
    controller:
      service:
        enabled: true
        type: LoadBalancer
EOF

helmrelease-wait-install ingress-traefik helm-release-ingress-traefik.yaml

echo ------------------------------------------
echo Installing Helm Chart kubernetes-dashboard
echo ------------------------------------------

cat > helm-release-kubernetes-dashboard.yaml <<EOF
apiVersion: app.undistro.io/v1alpha1
kind: HelmRelease
metadata:
  name: kubernetes-dashboard
  namespace: dashboard
spec:
  chart:
    repository: https://kubernetes.github.io/dashboard
    name: kubernetes-dashboard
    version: 3.0.2
  ## Namespace de instalacao do release
  ## Default=HelmRelease.metadata.namespace
  namespace: kube-system
  cluster:
    name: myeks
    ## Namespace do objeto Cluster no cluster Management
    ## Default=HelmRelease.metadata.namespace
    namespace: production-clusters
  autoUpgrade: true
  ## Lista de values, o ultimo sobrepões o anterior (merge)
  valuesFrom:
  - kind: ConfigMap
    name: dashboard-config
    valuesKeys: dashboard-config.yaml
  - kind: Secret
    name: dashboard-config-aws-secret
    valuesKeys: dashboard-config-aws-secrets.yaml
  ## Values inline. Sobrepões o resultado do merge do valuesFrom
  values:
    service:
      type: ClusterIP
    image:
      version: 1.0
  ## FIX: rodar antes de https://github.com/getupio-undistro/undistro/blob/master/controllers/app/helmrelease_controller.go#L197
  beforeApplyObjects:
  - apiVersion: v1
    kind: ConfigMap
    metadata:
      name: dashboard-config
      namespace: kubernetes-dashboard
    data:
      dashboard-config.yaml:
        ingress:
          enabled: true
        serviceAccount:
          name: undistro-dash
EOF

helmrelease-wait-install kubernetes-dashboard helm-release-kubernetes-dashboard.yaml

echo ------------------------------
echo Using EKS Worker Cluster
echo ------------------------------

undistro get kubeconfig -n default $WRK_CLUSTER > kubeconfig.$WRK_CLUSTER
export KUBECONFIG=$PWD/kubeconfig.$WRK_CLUSTER
helm ls -A
