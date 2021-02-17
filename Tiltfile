load('ext://restart_process', 'docker_build_with_restart')
load('ext://cert_manager', 'deploy_cert_manager')

IMG = 'controller:latest'
#docker_build(IMG, '.')

def yaml():
    return local('cd config/manager; kustomize edit set image controller=' + IMG + '; cd ../..; kustomize build config/default')

def manifests():
    return 'make manifests;'

def generate():
    return 'make generate;'

def vetfmt():
    return 'go vet ./...; go fmt ./...'

def binary():
    return 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o bin/manager main.go'

local(manifests() + generate())

deploy_cert_manager(version = 'v1.2.0')

local_resource('crd', manifests() + 'kustomize build config/crd | kubectl apply -f -', deps=['apis'])

#local_resource('un-crd', 'kustomize build config/crd | kubectl delete -f -', auto_init=False, trigger_mode=TRIGGER_MODE_MANUAL)

k8s_yaml(yaml())

local_resource('recompile', generate() + binary(), deps=['controllers', 'clustertemplates', 'main.go', 'pkg'])

local('kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/static/provider/kind/deploy.yaml')

docker_build_with_restart(IMG, '.', 
 dockerfile='tilt.docker',
 entrypoint='/manager',
 only=['./bin/manager', './clustertemplates'],
 live_update=[
       sync('./bin/manager', '/manager'),
       sync('./clustertemplates', '/clustertemplates'),
   ]
)