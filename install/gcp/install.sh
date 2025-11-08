#!/bin/bash


show_help () {
    echo "Usage: $0 [-h]"
    echo "  Deploy immotep on GCP"
    echo "      -h          : this message"
}
# Decode args
OPTIND=1  # Reset in case getopts has been used previously in the shell.
while getopts ":h?r:f" opt; do
    case "$opt" in
    h|\?)
        show_help
        exit 0
        ;;
    r)  
        export DOCKER_REGISTRY=${OPTARG}
        ;;
    f)  
        export FAST_BUILD=1
        ;;
    esac
done

# Create machine
ansible-playbook gcp_create.yml

# wait for machine avail

# ssh to try connect
vmip=$(head -2 gcpinv.ini | tail -1)

# wait 10x10s for VM to be ready
i=10
while [ $i != 0 ]; do
    i=$(( i - 1 ))

    down=$(nc -w 1 -z "${vmip}" 22 > /dev/null 2>&1; echo $?)

    if [ "${down}" != 0 ]; then
        echo "Wait for ${vmip} do be ready..."
        sleep 10
    else
        i=0
    fi
done

if [ "${down}" != 0 ]; then
    echo "Instance ${vmip} is not ready. Retry later."
    exit 1
else
    echo "Instance ${vmip} is accessible"
    # update ssh know_host

    res=$(grep -q "${vmip} " ~/.ssh/known_hosts; echo $?)
    if [ "${res}" == 1 ]; then
        echo "Adding ${vmip} to ssh known hosts"
        ssh-keyscan "${vmip}" >> ~/.ssh/known_hosts
    fi
fi

# install app

cd ../ansible || exit 1
# get vault password from gcp secret manager
gcloud secrets versions access latest --secret=ansible-vault > ansible_vlt_pwd
ansible-playbook --vault-password-file ansible_vlt_pwd -i ../gcp/gcpinv.ini immotep.yml
rm ansible_vlt_pwd
cd ../gcp || exit 1
