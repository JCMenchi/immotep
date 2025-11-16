#!/bin/bash


show_help () {
    echo "Usage: $0 [-h]"
    echo "  Deploy immotep on OVH"
    echo "      -h          : this message"
}
# Decode args
OPTIND=1  # Reset in case getopts has been used previously in the shell.
while getopts ":h?f" opt; do
    case "$opt" in
    h|\?)
        show_help
        exit 0
        ;;
    f)  
        export FAST_BUILD=1
        ;;
    esac
done

npm install

# set variables
PROJECT="Project 2025-10-22"
REGION=RBX-A
SSH_KEY_NAME="jcmate"
OS_IMAGE_NAME="Ubuntu 24.04"
FLAVOR_NAME="d2-2"
INSTANCE_NAME="immonode"
DNS_DOMAIN=jcm.ovh
SUB_DOMAIN="www"

# get IDs
PRJ_ID=$(ovhcloud cloud project list --format 'project_id+","+description' 2>/dev/null | grep "${PROJECT}" | cut -d',' -f1 | tr -d '"')
FLAVOR_ID=$(ovhcloud cloud reference list-flavors --cloud-project "${PRJ_ID}" --region "${REGION}"  --filter 'available && osType=="linux"' --format '"flavor"+name+","+id' 2>/dev/null | grep "flavor${FLAVOR_NAME}" | cut -d',' -f2 | tr -d '"')
IMAGE_ID=$(ovhcloud cloud reference list-images --cloud-project "${PRJ_ID}" --region "${REGION}" --filter "name==\"${OS_IMAGE_NAME}\" && status==\"active\" && type==\"linux\"" --format 'id' 2>/dev/null | tr -d '"')

#DNS_DOMAIN=jcm.ovh
SUB_DOMAIN="www"
# create instance
# 
INSTANCE_ID=$(ovhcloud cloud instance list --cloud-project "${PRJ_ID}" -f 'id+","+name' 2>/dev/null | grep "${INSTANCE_NAME}" | cut -d',' -f1 | tr -d '"')

if [ -n "${INSTANCE_ID}" ]; then
    echo "Instance ${INSTANCE_NAME} exists (${INSTANCE_ID})."
else
    echo "Creating instance ${INSTANCE_NAME}..."
    ovhcloud cloud instance create "${REGION}" --cloud-project "${PRJ_ID}" --name "${INSTANCE_NAME}" --flavor "${FLAVOR_ID}" --boot-from.image "${IMAGE_ID}" --network.public --ssh-key.name "${SSH_KEY_NAME}" 2>/dev/null 
    sleep 30
    INSTANCE_ID=$(ovhcloud cloud instance list --cloud-project "${PRJ_ID}" -f 'id+","+name' 2>/dev/null | grep "${INSTANCE_NAME}" | cut -d',' -f1 | tr -d '"')
    echo "Instance ${INSTANCE_NAME} created with ID (${INSTANCE_ID})."
fi

# get instance info
#IS_ACTIVE=$(ovhcloud cloud instance list --cloud-project "${PRJ_ID}" -f 'status' --filter "id==\"${INSTANCE_ID}\"" 2>/dev/null | tr -d '"')

VM_PUBLLIC_IP=$(ovhcloud cloud instance  get $INSTANCE_ID  --json 2>/dev/null | jq '.["ipAddresses"][] | select( .version == 4 and .type == "public" ) .ip' | tr -d '"')
# wait 10x10s for VM to be ready
i=10
down=0
while [ $i != 0 ]; do
    i=$(( i - 1 ))

    INSTANCE_ID=$(ovhcloud cloud instance list --cloud-project "${PRJ_ID}" -f 'id+","+name' 2>/dev/null | grep "${INSTANCE_NAME}" | cut -d',' -f1 | tr -d '"')
    VM_PUBLLIC_IP=$(ovhcloud cloud instance  get $INSTANCE_ID  --json 2>/dev/null | jq '.["ipAddresses"][] | select( .version == 4 and .type == "public" ) .ip' | tr -d '"')
    down=$(nc -w 1 -z "${VM_PUBLLIC_IP}" 22 > /dev/null 2>&1; echo $?)

    if [ "${down}" != 0 ]; then
        echo "Wait for ${INSTANCE_ID} ${VM_PUBLLIC_IP} do be ready..."
        sleep 10
    else
        i=0
    fi
done

if [ "${down}" != 0 ]; then
    echo "Instance ${VM_PUBLLIC_IP} is not ready. Retry later."
    exit 1
fi
echo "Instance ${VM_PUBLLIC_IP} is accessible"

# check if IP is in DNS

CURRENT_IP=$(ovhcloud domain-zone get "${DNS_DOMAIN}" --json | jq ".[\"records\"][] | select(.fieldType == \"A\" and .subDomain==\"${SUB_DOMAIN}\").target" | tr -d '"')

if [ "${CURRENT_IP}" != "${VM_PUBLLIC_IP}" ]; then
    echo "Updating DNS record for ${SUB_DOMAIN}.${DNS_DOMAIN} to ${VM_PUBLLIC_IP} (was ${CURRENT_IP})"
    RECORD_ID=$(ovhcloud domain-zone get "${DNS_DOMAIN}" --json | jq ".[\"records\"][] | select(.fieldType == \"A\" and .subDomain==\"${SUB_DOMAIN}\").id" | tr -d '"')
    if [ -z "${RECORD_ID}" ]; then
        echo "DNS record for ${SUB_DOMAIN}.${DNS_DOMAIN} does not exist. Creating it."
        node add_dns_record.js "${DNS_DOMAIN}" "${SUB_DOMAIN}" "${VM_PUBLLIC_IP}"
    else
        ovhcloud domain-zone record update "${DNS_DOMAIN}" "${RECORD_ID}" --target "${VM_PUBLLIC_IP}"
    fi
   
    ovhcloud domain-zone refresh "${DNS_DOMAIN}"
else
    echo "DNS record for ${SUB_DOMAIN}.${DNS_DOMAIN} is up to date (${CURRENT_IP})"
fi

# update ssh know_host
res=$(grep -q "${VM_PUBLLIC_IP} " ~/.ssh/known_hosts; echo $?)
if [ "${res}" == 1 ]; then
    echo "Adding ${VM_PUBLLIC_IP} to ssh known hosts"
    ssh-keyscan "${VM_PUBLLIC_IP}" >> ~/.ssh/known_hosts
fi

# update inventory file
echo "[apphosts]" > ovhinv.ini
echo "${VM_PUBLLIC_IP}" >> ovhinv.ini

echo "[apphosts:vars]" >> ovhinv.ini
echo "ansible_ssh_user=ubuntu" >> ovhinv.ini
echo "ansible_ssh_private_key_file=~/.ssh/ovh" >> ovhinv.ini

# install app
cd ../ansible || exit 1

ansible-playbook --vault-password-file ansible_vlt_pwd -e email_for_certbot="${CERTBOT_EMAIL}" -e web_server_url=${SUB_DOMAIN}.${DNS_DOMAIN} -i ../ovh/ovhinv.ini immotep.yml

cd ../ovh || exit 1

echo "Deployment completed. Access your app at https://${SUB_DOMAIN}.${DNS_DOMAIN}/"
