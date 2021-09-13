#!/bin/bash

# Create group if needed
r=$(getent group immotep >/dev/null 2>&1; echo $?)
if [ "$r" ]; then
    groupadd immotep
fi

# create user of needed
r=$(getent passwd immotep >/dev/null 2>&1; echo $?)
if [ "$r" ]; then
    useradd -g immotep -s /sbin/nologin immotep
fi

systemctl stop immotep.service

cp immotep.env /etc/default/immotep
cp immotep.service /etc/systemd/system
cp ../../srv/immotep /usr/bin

systemctl enable immotep.service
systemctl start immotep.service
