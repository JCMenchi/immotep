#!/bin/bash

export DOCKER_REGISTRY=${OCI_DOCKER_REGISTRY:-localhost:5000/}

# Create Postgresql service
(
    cd pgsql || exit

    docker build -t "${DOCKER_REGISTRY}"pgsql:13 .
    docker push "${DOCKER_REGISTRY}"pgsql:13

    PG_ADMIN_USER=${PG_ADMIN_USER:-pgroot}
    PG_ADMIN_PASSWORD=$(tr -dc _A-Z-a-z-0-9 < /dev/urandom | head -c${1:-32})

    # create secrets if needed
    n=$(kubectl get secrets | grep -q pgsql-admin; echo $?)
    if [ "$n" == 1 ]; then
        kubectl create secret generic pgsql-admin --from-literal=username="${PG_ADMIN_USER}" --from-literal=password="${PG_ADMIN_PASSWORD}"
    fi

    kubectl apply -f pg_volume.yaml
    kubectl apply -f pg_service.yaml
)

# wait for pgsql to be ready
pgpod=$(kubectl get pods --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | grep pgsql)
kubectl wait --for=condition=ready pod/"${pgpod}"
# rem wait a few second for postgresql init
sleep 10

# create db for backend
PG_IMMOTEPDB_USER=${PG_IMMOTEPDB_USER:-immuser}
PG_IMMOTEPDB_PASSWORD=$(tr -dc _A-Z-a-z-0-9 < /dev/urandom | head -c${1:-32})

kubectl exec "${pgpod}" -- /run/postgresql/pg_add_app.sh -u "${PG_IMMOTEPDB_USER}" -p "${PG_IMMOTEPDB_PASSWORD}" immotep
