load('ext://restart_process', 'docker_build_with_restart')


IMG = 'controller:latest'

def yaml():
    return local('cd config/manager; kustomize edit set image controller=' + IMG + '; cd ../..; kustomize build config/default')

def manifests():
    return 'make generate-manifests;'

def generate():
    return 'make generate-go-core;'

def vetfmt():
    return 'go vet ./...; go fmt ./...'

def binary():
    return 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o bin/manager main.go'

def init():
    print('Init undistro')
    return local('make undistro && ./bin/undistro --config undistro.yaml init')

init()

local(manifests() + generate())

local_resource('crd', manifests() + 'kustomize build config/crd | kubectl apply -f -', deps=["api"])

k8s_yaml(yaml())

local_resource('recompile', generate() + binary(), deps=['controllers', 'client', 'templates', 'main.go'])

docker_build_with_restart(IMG, '.', 
 dockerfile='tilt.docker',
 entrypoint='/app/manager',
 only=['./bin/manager', './clustertemplates'],
 live_update=[
       sync('./bin/manager', '/app/manager'),
       sync('./clustertemplates', '/app/clustertemplates'),
   ]
)