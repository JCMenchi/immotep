#!/bin/bash


# Get info about region, department, city

# Official zipcode
# readable URL is
# https://datanova.legroupe.laposte.fr/explore/dataset/laposte_hexasmal/download/?format=csv&timezone=Europe/Berlin&use_labels_for_header=true

if [ ! -f zip_code.csv ]; then
    curl 'https://datanova.legroupe.laposte.fr/explore/dataset/laposte_hexasmal/download/?format=csv&timezone=Europe/Berlin&use_labels_for_header=true' > zip_code.csv
fi

#
# Use https://geo.api.gouv for basic administrative info
#
# Use laposte for geojson info
#  https://datanova.legroupe.laposte.fr/api/records/1.0/geopreview/?disjunctive.reg_name=true&sort=year&rows=1000&clusterprecision=6&dataset=georef-france-region&timezone=Europe%2FBerlin
# 
if [ ! -f regions.geojson ]; then
    curl 'https://datanova.legroupe.laposte.fr/explore/dataset/georef-france-region/download/?format=geojson&timezone=Europe/Berlin' > regions.geojson
fi

if [ ! -f departements.geojson ]; then
    curl 'https://datanova.legroupe.laposte.fr/explore/dataset/georef-france-departement/download/?format=geojson&timezone=Europe/Berlin' > departements.geojson
fi

if [ ! -f communes.json ]; then
    curl 'https://geo.api.gouv.fr/communes' > communes.json
fi

#
# Get sales infos from DVF database
# https://www.data.gouv.fr/fr/datasets/demandes-de-valeurs-foncieres/
#

DVF_BASE_URL=https://static.data.gouv.fr/resources/demandes-de-valeurs-foncieres/20210330-

if [ ! -f valeursfoncieres-2020.txt ]; then
    curl ${DVF_BASE_URL}102242/valeursfoncieres-2020.txt > valeursfoncieres-2020.txt
fi

if [ ! -f valeursfoncieres-2019.txt ]; then
    curl ${DVF_BASE_URL}102025/valeursfoncieres-2019.txt > valeursfoncieres-2019.txt
fi

if [ ! -f valeursfoncieres-2018.txt ]; then
    curl ${DVF_BASE_URL}101812/valeursfoncieres-2018.txt > valeursfoncieres-2018.txt
fi

if [ ! -f valeursfoncieres-2017.txt ]; then
    curl ${DVF_BASE_URL}101548/valeursfoncieres-2017.txt > valeursfoncieres-2017.txt
fi

if [ ! -f valeursfoncieres-2016.txt ]; then
    curl ${DVF_BASE_URL}101325/valeursfoncieres-2016.txt > valeursfoncieres-2016.txt
fi

# build exe
if [ ! -f immotep ]; then
    if [ ! -d ../srv/api/immotep ]; then
        if [ ! -d ../ui/immotep ]; then
            (cd ../ui || return; npm install; npm run build:dist)
        fi
        mv ../ui/immotep ../srv/api
    fi
    
    (cd ../srv || return; go build)
    mv ../srv/immotep .
fi

# Build database from raw data
if [ ! -f imm.db ]; then
    ./immotep -f imm.db loadconf --region regions.geojson --department departements.geojson --city communes.json
    ./immotep -f imm.db load -z zip_code.csv valeursfoncieres-20*.txt
    ./immotep -f imm.db geocode
    ./immotep -f imm.db compute
fi
