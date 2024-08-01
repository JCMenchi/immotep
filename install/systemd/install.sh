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

setcap 'cap_net_bind_service=+ep' /usr/bin/immotep

# create database
su postgres -c "psql -dpostgres -c \"CREATE USER immotep WITH PASSWORD 'imm.pwd12'\" "
su postgres -c "psql -dpostgres -c \"CREATE DATABASE immdb WITH OWNER = 'immotep'\" "
su postgres -c "psql -dpostgres -c \"GRANT ALL PRIVILEGES ON DATABASE immdb to immotep;\" "


systemctl enable immotep.service
systemctl start immotep.service
