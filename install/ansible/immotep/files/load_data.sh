#!/bin/bash

# get service configuration
set -a # export all var
# shellcheck disable=SC1091
. /etc/default/immotep.env


GEOFILES="communes.geojson communes.json departements.geojson regions.geojson"

# start with geo data
if [ -f /data/communes.geojson.xz ] && [ -f /data/regions.geojson.xz ] && [ -f /data/departements.geojson.xz ] && [ -f /data/communes.json.xz ]; then
    if [ ! -d /data/geo ]; then
        mkdir /data/geo
    fi

    cd /data/geo || exit 1

    for fi in ${GEOFILES}; do
        if [ -f ../"${fi}".xz ]; then
            xz -c -d ../"${fi}".xz > "${fi}"
        fi
    done

    if [ -f communes.geojson ] && [ -f regions.geojson ] && [ -f departements.geojson ] && [ -f communes.json ]; then
        immotep loadconf --region regions.geojson --department departements.geojson --city communes.json --citygeo communes.geojson
    fi
fi

# Load data
if [ ! -d /data/value ]; then
    mkdir /data/value
fi
cd /data/value || exit 2
YEARS="2016 2017 2018 2019 2020 2021 2022 2023"
for y in ${YEARS}; do
    if [ ! -f valeursfoncieres-"${y}".txt.loaded ]; then
        if [ -f ../valeursfoncieres-"${y}".txt.xz ]; then
            xz -c -d ../valeursfoncieres-"${y}".txt.xz > valeursfoncieres-"${y}".txt
            immotep load valeursfoncieres-"${y}".txt
            touch valeursfoncieres-"${y}".txt.loaded
            immotep geocode
        fi
    fi
done

immotep geocode
# compute stat
immotep compute
immotep aggregate