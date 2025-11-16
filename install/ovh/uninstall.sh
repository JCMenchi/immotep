#!/bin/bash


show_help () {
    echo "Usage: $0 [-h]"
    echo "  Suppress immotep server from OVH"
    echo "      -h          : this message"
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


# set variables
PROJECT="Project 2025-10-22"
INSTANCE_NAME="immonode"
DNS_DOMAIN=jcm.ovh
SUB_DOMAIN="www"

# get IDs
PRJ_ID=$(ovhcloud cloud project list --format 'project_id+","+description' 2>/dev/null | grep "${PROJECT}" | cut -d',' -f1 | tr -d '"')

SUB_DOMAIN="www"

# install ovhcli js
npm install

# get instance
# 
INSTANCE_ID=$(ovhcloud cloud instance list --cloud-project "${PRJ_ID}" -f 'id+","+name' 2>/dev/null | grep "${INSTANCE_NAME}" | cut -d',' -f1 | tr -d '"')

if [ -n "${INSTANCE_ID}" ]; then
    echo "Instance ${INSTANCE_NAME} exists (${INSTANCE_ID})."
    VM_PUBLLIC_IP=$(ovhcloud cloud instance  get "${INSTANCE_ID}"  --json 2>/dev/null | jq '.["ipAddresses"][] | select( .version == 4 and .type == "public" ) .ip' | tr -d '"')

    # check if IP is in DNS
    CURRENT_IP=$(ovhcloud domain-zone get "${DNS_DOMAIN}" --json | jq ".[\"records\"][] | select(.fieldType == \"A\" and .subDomain==\"${SUB_DOMAIN}\").target" | tr -d '"')

    if [ "${CURRENT_IP}" != "${VM_PUBLLIC_IP}" ]; then
        echo "DNS use another IP (${CURRENT_IP} != ${VM_PUBLLIC_IP}). Nothing to do."
    else
        echo "DNS record for ${SUB_DOMAIN}.${DNS_DOMAIN} is up to date (${CURRENT_IP})"
        echo "Deleting DNS record for ${SUB_DOMAIN}.${DNS_DOMAIN} ..."
        RECORD_ID=$(ovhcloud domain-zone get "${DNS_DOMAIN}" --json | jq ".[\"records\"][] | select(.fieldType == \"A\" and .subDomain==\"${SUB_DOMAIN}\").id")
        
        node del_dns_record.js "${RECORD_ID}"

        echo "DNS record deleted."
    fi

    echo "Deleting instance ${INSTANCE_NAME} (${INSTANCE_ID}) ..."
    ovhcloud cloud instance delete "${INSTANCE_ID}"
else
    echo "Instance ${INSTANCE_NAME} does not exist."
fi

# update inventory file
if [ -f ovhinv.ini ]; then
    rm ovhinv.ini
fi


# https://www.ovh.com/engine/2api/sws/domain/zone/jcm.ovh/records
# {"records":[5389566490]}

#  curl -X DELETE "https://eu.api.ovh.com/v1/domain/zone/${DNS_DOMAIN}/record/${RECORD_ID}" \


# 
# TOKEN=$(curl --request POST \
#    --url 'https://www.ovh.com/auth/oauth2/token' \
#    --header 'content-type: application/x-www-form-urlencoded' \
#    --data grant_type=client_credentials \
#    --data client_id=${OVH_OAUTH2_CLIENT_ID} \
#    --data client_secret=${OVH_OAUTH2_CLIENT_SECRET} \
#    --data scope=all | cut -d \" -f 4)
# 
# curl -H "Authorization:Bearer ${TOKEN}" "https://eu.api.ovh.com/v1/domain/zone/jcm.ovh/record/"