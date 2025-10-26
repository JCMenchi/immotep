#!/bin/bash

set +x
export VERSION=0.0.1
VERSION=$(grep version ui/package.json | cut -d\" -f4)
export DOCKER_REGISTRY=${DOCKER_REGISTRY:-localhost:5000}
export FAST_BUILD=0
export DOCKER_BUILD=0

show_help () {
    echo "Usage: $0 [-h] [-f] [-c] [-r registry]"
    echo "  Build podman file"
    echo "      -r registry : podman registry (default: ${DOCKER_REGISTRY})"
    echo "      -f          : fast build"
}
# Decode args
OPTIND=1  # Reset in case getopts has been used previously in the shell.
while getopts ":h?r:fc" opt; do
    case "$opt" in
    h|\?)
        show_help
        exit 0
        ;;
    r)  
        export DOCKER_REGISTRY=${OPTARG}
        ;;
    f)  
        export FAST_BUILD=1
        ;;
    c)  
        export DOCKER_BUILD=1
        ;;
    esac
done

if [ "${FAST_BUILD}" != 1 ]; then
    echo "================ CLEAN ========================="
    echo "Remove ui/dist"
    rm -rf ./ui/dist
    echo "Remove srv/api/immotep"
    rm -rf ./srv/api/immotep
    echo "Remove ui/node_modules"
    rm -rf ./ui/node_modules
    rm -rf ./ui/package-lock.json
fi

# Build UI application
(
    cd ui || exit 1
    echo "================ UI ========================="
    if [ ! -d node_modules ]; then
        echo "Instal node modules"
        npm install
        if [ -d dist ]; then
            rm -rf dist
        fi
    fi
    if [ ! -d dist ]; then
        echo "Build UI"
        npm run build
        if [ -d ../srv/api/immotep ]; then
            rm -rf ../srv/api/immotep
        fi
    fi
    if [ ! -d ../srv/api/immotep ]; then
        echo "Copy UI to backend"
        mkdir -p ../srv/api/immotep
        cp -r dist/* ../srv/api/immotep
    fi
)

# Build Backend application
(
    cd srv || exit 1
    echo "================ BACKEND ========================="
    echo "Build backend"
    go build -o immotep ./main.go
)

if [ "${DOCKER_BUILD}" == 1 ]; then
    echo "================ BACKEND ========================="
    # Package as container
    podman build -t immotep:"${VERSION}" .

    if [ "${DOCKER_REGISTRY}X" != "X" ]; then
        echo "Deploy to registry: ${DOCKER_REGISTRY}"
        podman tag immotep:"${VERSION}" "${DOCKER_REGISTRY}"immotep:"${VERSION}"
        podman push "${DOCKER_REGISTRY}"immotep:"${VERSION}"

        echo "================ K8S ========================="
        kubectl apply -f k8s_deploy.yaml
    fi
fi
echo "================ DONE ========================="