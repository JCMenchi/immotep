#!/bin/bash

set +x
export VERSION=0.0.1
VERSION=$(grep version ui/package.json | cut -d\" -f4)
export DOCKER_REGISTRY=${DOCKER_REGISTRY:-localhost:5000}
export FAST_BUILD=0

show_help () {
    echo "Usage: $0 [-h] [-f] [-r registry]"
    echo "  Build docker file"
    echo "      -r registry : docker registry (default: ${DOCKER_REGISTRY})"
    echo "      -f          : fast build"
}
# Decode args
OPTIND=1  # Reset in case getopts has been used previously in the shell.
while getopts ":h?r:f" opt; do
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
    esac
done

if [ "${FAST_BUILD}" != 1 ]; then
    echo "================ CLEAN ========================="
    echo "Remove ui/immotep"
    rm -rf ./ui/immotep
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
        npm install --no-optional --no-package-lock
        if [ -d immotep ]; then
            rm -rf immotep
        fi
    fi
    if [ ! -d immotep ]; then
        echo "Build UI"
        npm run build:dist
        if [ -d ../srv/api/immotep ]; then
            rm -rf ../srv/api/immotep
        fi
    fi
    if [ ! -d ../srv/api/immotep ]; then
        echo "Copy UI to backend"
        cp -r immotep ../srv/api/immotep
    fi
)

echo "================ BACKEND ========================="
# Package as container
docker build -t immotep:"${VERSION}" .

if [ "${DOCKER_REGISTRY}X" != "X" ]; then
    echo "Deploy to registry: ${DOCKER_REGISTRY}"
    docker tag immotep:"${VERSION}" "${DOCKER_REGISTRY}"immotep:"${VERSION}"
    docker push "${DOCKER_REGISTRY}"immotep:"${VERSION}"

    echo "================ K8S ========================="
    kubectl apply -f k8s_deploy.yaml
fi
