#!/bin/bash

set +x
export VERSION=0.0.1
VERSION=$(grep version ui/package.json | cut -d\" -f4)
export DOCKER_REGISTRY=${DOCKER_REGISTRY:-localhost:5000}

show_help () {
    echo "Usage: $0 [-h] [-i ingress] [-r registry]"
    echo "  Build docker file"
    echo "      -r registry : docker registry (default: ${DOCKER_REGISTRY})"
}
# Decode args
OPTIND=1  # Reset in case getopts has been used previously in the shell.
while getopts ":h?r:" opt; do
    case "$opt" in
    h|\?)
        show_help
        exit 0
        ;;
    r)  
        export DOCKER_REGISTRY=${OPTARG}
        ;;
    esac
done

# Build UI application
(
    cd ui || exit 1
    npm install --no-optional --no-package-lock
    npm run build
)

# Package as container
docker build -t immotep:"${VERSION}" .

if [ "${DOCKER_REGISTRY}X" != "X" ]; then
    echo "Deploy to registry: ${DOCKER_REGISTRY}"
    docker tag immotep:"${VERSION}" "${DOCKER_REGISTRY}"immotep:"${VERSION}"
    docker push "${DOCKER_REGISTRY}"immotep:"${VERSION}"
fi


