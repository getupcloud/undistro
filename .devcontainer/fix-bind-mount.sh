#!/bin/bash
if [ "${CODESPACES}" != "true" ]; then
    echo 'Not in a codespace. Aborting.'
    exit 0
fi

WORKSPACE_PATH_IN_CONTAINER=${1:-"$HOME/workspace"}
shift
WORKSPACE_PATH_ON_HOST=${2:-"/var/lib/docker/vsonlinemount/workspace"}
shift
VM_CONTAINER_WORKSPACE_PATH=/vm-host/$WORKSPACE_PATH_IN_CONTAINER
VM_CONTAINER_WORKSPACE_BASE_FOLDER=$(dirname $VM_CONTAINER_WORKSPACE_PATH)
VM_HOST_WORKSPACE_PATH=/vm-host/$WORKSPACE_PATH_ON_HOST

echo -e "Workspace path in container: ${WORKSPACE_PATH_IN_CONTAINER}\nWorkspace path on host: ${WORKSPACE_PATH_ON_HOST}"
docker run --rm -v /:/vm-host alpine sh -c "\
    if [ -d "${VM_CONTAINER_WORKSPACE_PATH}" ]; then echo \"${WORKSPACE_PATH_IN_CONTAINER} already exists on host. Aborting.\" && return 0; fi
    apk add coreutils > /dev/null \
    && mkdir -p $VM_CONTAINER_WORKSPACE_BASE_FOLDER \
    && cd $VM_CONTAINER_WORKSPACE_BASE_FOLDER \
    && ln -s \$(realpath --relative-to='.' $VM_HOST_WORKSPACE_PATH) .\
    && echo 'Symlink created!'"