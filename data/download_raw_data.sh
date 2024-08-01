#!/bin/bash


# Get info about region, department, city


#
# Use https://geo.api.gouv for basic administrative info
#
if [ ! -f communes.json ]; then
    curl 'https://geo.api.gouv.fr/communes' > communes.json
fi

#
# Get sales infos from DVF database
# https://www.data.gouv.fr/fr/datasets/demandes-de-valeurs-foncieres/
#

DVF_BASE_URL=https://static.data.gouv.fr/resources/demandes-de-valeurs-foncieres/20240408-

if [ ! -f valeursfoncieres-2023.txt ]; then
    # static url https://www.data.gouv.fr/fr/datasets/r/78348f03-a11c-4a6b-b8db-2acf4fee81b1
    curl ${DVF_BASE_URL}125738/valeursfoncieres-2023.txt > valeursfoncieres-2023.txt
fi

if [ ! -f valeursfoncieres-2022.txt ]; then
    # curl https://static.data.gouv.fr/resources/demandes-de-valeurs-foncieres/20221017-152027/valeursfoncieres-2022-s1.txt > valeursfoncieres-2022.txt
    curl ${DVF_BASE_URL}130630/valeursfoncieres-2022.txt > valeursfoncieres-2022.txt
    # static url https://www.data.gouv.fr/fr/datasets/r/87038926-fb31-4959-b2ae-7a24321c599a
fi

if [ ! -f valeursfoncieres-2021.txt ]; then
    #curl https://static.data.gouv.fr/resources/demandes-de-valeurs-foncieres/20221017-151704/valeursfoncieres-2021.txt > valeursfoncieres-2021.txt
    curl ${DVF_BASE_URL}130153/valeursfoncieres-2021.txt > valeursfoncieres-2021.txt
    # static url https://www.data.gouv.fr/fr/datasets/r/817204ac-2202-4b4a-98e7-4184d154d98c
fi

if [ ! -f valeursfoncieres-2020.txt ]; then
    #curl ${DVF_BASE_URL}102242/valeursfoncieres-2020.txt > valeursfoncieres-2020.txt
    curl ${DVF_BASE_URL}125058/valeursfoncieres-2020.txt > valeursfoncieres-2020.txt
    # static url https://www.data.gouv.fr/fr/datasets/r/90a98de0-f562-4328-aa16-fe0dd1dca60f
fi

if [ ! -f valeursfoncieres-2019.txt ]; then
    #curl ${DVF_BASE_URL}102025/valeursfoncieres-2019.txt > valeursfoncieres-2019.txt
    curl ${DVF_BASE_URL}124817/valeursfoncieres-2019.txt > valeursfoncieres-2019.txt
    # static url https://www.data.gouv.fr/fr/datasets/r/3004168d-bec4-44d9-a781-ef16f41856a2
fi

DVF_BASE_URL=https://static.data.gouv.fr/resources/demandes-de-valeurs-foncieres/20210330-
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
    ./immotep -f imm.db loadconf --region regions.geojson --department departements.geojson --city communes.json --citygeo communes.geojson
    ./immotep -f imm.db load valeursfoncieres-20*.txt
    ./immotep -f imm.db geocode
    ./immotep -f imm.db compute
fi

./immotep loadconf --region ../data/regions.geojson --department ../data/departements.geojson --city ../data/communes.json --citygeo ../data/communes.geojson
./immotep load ../data/valeursfoncieres-small.txt
./immotep compute

./immotep loadconf --region data/regions.geojson --department data/departements.geojson --city data/communes.json --citygeo data/communes.geojson
./immotep load ../data/valeursfoncieres-small.txt
./immotep compute
