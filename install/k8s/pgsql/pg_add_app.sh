#!/bin/bash

PG_HOME=/usr
PGHOST=${PGHOST:-127.0.0.1}
PGPORT=${PGPORT:-5432}

PG_ADMIN_USER=${PG_ADMIN_USER:-pgr}
PG_ADMIN_PASSWORD=${PG_ADMIN_PASSWORD:-pgr}

export USE_K8S=0
# generate random values for default
APP_USER=$(tr -cd '[:lower:]' < /dev/urandom | fold -w16 | head -n1)
APP_PASSWORD=$(tr -cd '[:alnum:]' < /dev/urandom | fold -w30 | head -n1)


show_help () {
    echo "Usage: $0 [-h] [ -u username] [ -p password ] [-k] dbname"
    echo "  Create new Database"
    echo "      -u username : database user name"
    echo "      -p password : database user password"
    echo "      -k : add k8s secret"
}
# Decode args
OPTIND=1  # Reset in case getopts has been used previously in the shell.
while getopts ":h?u:p:k" opt; do
    case "$opt" in
    h|\?)
        show_help
        exit 0
        ;;
    u)
        export APP_USER=${OPTARG}
        ;;
    p)
        export APP_PASSWORD=${OPTARG}
        ;;
    k)
        export USE_K8S=1
        ;;
    esac
done
shift $((OPTIND -1))

if [ $# -ne 1 ]; then
	echo "Wrong number of argument."
	show_help
	exit 1
fi

APP=$1

APP_DB=${APP}db

# check if database exists
n=$(PGPASSWORD=${PG_ADMIN_PASSWORD} psql -U "${PG_ADMIN_USER}" -h"${PGHOST}" -p"${PGPORT}" -dpostgres -c'\l' | grep -c "${APP_DB}" )
if [ "${n}" -ne 0 ]; then
	echo "Database ${APP_DB} already created."
	exit 2
fi

# check if user exists
n=$(PGPASSWORD=${PG_ADMIN_PASSWORD} psql -U "${PG_ADMIN_USER}" -h"${PGHOST}" -p"${PGPORT}" -dpostgres -c 'select rolname FROM pg_roles;' | grep -c "${APP_USER}" )
if [ "${n}" -eq 0 ]; then
	echo "Create user: ${APP_USER}"
	PGPASSWORD=${PG_ADMIN_PASSWORD} "${PG_HOME}"/bin/psql -U "${PG_ADMIN_USER}" -h"${PGHOST}" -p"${PGPORT}" -dpostgres -c "CREATE USER ${APP_USER} WITH ROLE ${PG_ADMIN_USER} PASSWORD '${APP_PASSWORD}'"
	PGPASSWORD=${PG_ADMIN_PASSWORD} "${PG_HOME}"/bin/psql -U "${PG_ADMIN_USER}" -h"${PGHOST}" -p"${PGPORT}" -dpostgres -c "CREATE DATABASE ${APP_DB} WITH OWNER = '${APP_USER}'"
	PGPASSWORD=${PG_ADMIN_PASSWORD} "${PG_HOME}"/bin/psql -U "${PG_ADMIN_USER}" -h"${PGHOST}" -p"${PGPORT}" -dpostgres -c "GRANT ALL PRIVILEGES ON DATABASE ${APP_DB} to ${APP_USER};"
else
	echo "User ${APP_USER} already created."
	exit 3
fi

# execute sql script with db name if provided
if [ -e "${APP}".sql ]; then
	PGPASSWORD=${APP_PASSWORD} "${PG_HOME}"/bin/psql -U "${APP_USER}" -h"${PGHOST}"  -p"${PGPORT}" -d"${APP_DB}" -f "${APP}".sql
fi

if [ $USE_K8S == 1 ]; then
# Add secret for service
    n=$(kubectl get secrets | grep -q "${APP_DB}"-admin; echo $?)
    if [ "$n" == 1 ]; then
        kubectl create secret generic "${APP}"-pgsql \
                --from-literal=username="${APP_USER}" \
                --from-literal=password="${APP_PASSWORD}" \
                --from-literal=database="${APP_DB}" \
                --from-literal=dbhost="pgsql"
    fi
fi

# show users
PGPASSWORD=${PG_ADMIN_PASSWORD} "${PG_HOME}"/bin/psql -U "${PG_ADMIN_USER}" -h"${PGHOST}"  -p"${PGPORT}" -dpostgres -c'\du'
# show databases
PGPASSWORD=${PG_ADMIN_PASSWORD} "${PG_HOME}"/bin/psql -U "${PG_ADMIN_USER}" -h"${PGHOST}"  -p"${PGPORT}" -dpostgres -c'\l'

if [ $USE_K8S == 0 ]; then
    # generate yaml config
    echo "# add the following DB conf to ~/.immotep.yaml"
    echo "dsn:"
    echo "  type: pgsql"
    echo "  host: ${PGHOST}"
    echo "  port: ${PGPORT}"
    echo "  dbname: ${APP_DB}"
    echo "  user: ${APP_USER}"
    echo "  password: ${APP_PASSWORD}"
fi