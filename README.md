# Data collection

Pour les données de base DVF
<https://www.data.gouv.fr/fr/datasets/demandes-de-valeurs-foncieres/>

Pour le geocodage:
<https://www.data.gouv.fr/fr/datasets/base-adresse-nationale/>

Pour le code postaux manquants
<https://www.data.gouv.fr/fr/datasets/base-officielle-des-codes-postaux/>

Pour les lignes haute tension
<https://opendata.reseaux-energies.fr/explore/dataset/lignes-aeriennes-rte/information/?disjunctive.etat&disjunctive.tension>

Pour les limites departements et ville
<https://geo.api.gouv.fr/decoupage-administratif/communes>

<https://geo.api.gouv.fr/communes?lat=48.838052499999996&lon=2.7151414&fields=code,nom,codesPostaux,surface,population,centre,contour>

Get all regions
curl 'https://geo.api.gouv.fr/regions'
Get all dep for a region
curl 'https://geo.api.gouv.fr/regions/{regioncode}/departement'
Get all communes for a departement
curl 'https://geo.api.gouv.fr/departements/{depcode}/communes?fields=center'

curl 'https://geo.api.gouv.fr/communes/{codecommune}?fields=code,nom,contour'
