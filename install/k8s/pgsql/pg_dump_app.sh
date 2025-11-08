#!/bin/bash

PGHOST=${PGHOST:-127.0.0.1}
PGPORT=${PGPORT:-5432}

show_help () {
    echo "Usage: $0 [-h] dbname"
    echo "  Dump Database"
}
# Decode args
OPTIND=1  # Reset in case getopts has been used previously in the shell.
while getopts ":h?" opt; do
    case "$opt" in
    h|\?)
        show_help
        exit 0
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
APP_USER=${APP}user

PGPASSWORD=${APP_USER} pg_dump --dbname="${APP_DB}" -U"${APP_USER}" -p"${PGPORT}" --create -h"${PGHOST}"
