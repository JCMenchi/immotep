#!/bin/bash

PG_HOME=/usr
PGHOST=${PGHOST:-127.0.0.1}
PGPORT=${PGPORT:-5432}

PG_ADMIN_USER=${PG_ADMIN_USER:-pgr}
PG_ADMIN_PASSWORD=${PG_ADMIN_PASSWORD:-pgr}

export USE_K8S=0

show_help () {
    echo "Usage: $0 [-h] dbname"
    echo "  Delete Database"
}
# Decode args
OPTIND=1  # Reset in case getopts has been used previously in the shell.
while getopts ":h?ku:" opt; do
    case "$opt" in
    h|\?)
        show_help
        exit 0
        ;;
    u)
        export APP_USER=${OPTARG}
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


if [ $USE_K8S == 1 ]; then
    # check if secret 
    n=$(kubectl get secrets | grep -q "${APP}"-pgsql; echo $?)
    if [ "$n" == 1 ]; then
        echo "Cannot find secret ${APP}-pgsql"
        exit 2
    fi

    # get value from secret
    APP_DB=$(kubectl get secret "${APP}"-pgsql -o yaml | grep " database:" | cut -d: -f2 | cut -d\  -f2 | base64 -d)
    APP_USER=$(kubectl get secret "${APP}"-pgsql -o yaml | grep " username:" | cut -d: -f2 | cut -d\  -f2 | base64 -d)
fi

# check if database exists
n=$(PGPASSWORD=${PG_ADMIN_PASSWORD} psql -U "${PG_ADMIN_USER}" -h"${PGHOST}" -p"${PGPORT}" -dpostgres -c'\l' | grep -c "${APP_DB}" )
if [ "${n}" -eq 0 ]; then
	echo "Database ${APP_DB} does not exist."
	exit 3
else
    PGPASSWORD=${PG_ADMIN_PASSWORD} ${PG_HOME}/bin/psql -U "${PG_ADMIN_USER}" -h"${PGHOST}" -p"${PGPORT}" -dpostgres -c "DROP DATABASE ${APP_DB};"
fi

# check if user exists
n=$(PGPASSWORD=${PG_ADMIN_PASSWORD} psql -U "${PG_ADMIN_USER}" -h"${PGHOST}" -p"${PGPORT}" -dpostgres -c 'select rolname FROM pg_roles;' | grep -c "${APP_USER}" )
if [ "${n}" -eq 0 ]; then
	echo "User ${APP_USER} does not exist."
	exit 4
else
	PGPASSWORD=${PG_ADMIN_PASSWORD} ${PG_HOME}/bin/psql -U "${PG_ADMIN_USER}" -h"${PGHOST}" -p"${PGPORT}" -dpostgres -c "DROP USER ${APP_USER};"
fi

if [ $USE_K8S == 1 ]; then
    # delete secret
    kubectl delete secrets "${APP}"-pgsql
fi

# show users
PGPASSWORD=${PG_ADMIN_PASSWORD} "${PG_HOME}"/bin/psql -U "${PG_ADMIN_USER}" -h"${PGHOST}"  -p"${PGPORT}" -dpostgres -c'\du'
# show databases
PGPASSWORD=${PG_ADMIN_PASSWORD} "${PG_HOME}"/bin/psql -U "${PG_ADMIN_USER}" -h"${PGHOST}"  -p"${PGPORT}" -dpostgres -c'\l'
