#!/usr/bin/env node

const ovh = require('@ovhcloud/node-ovh')({
  appKey: process.env.OVH_APPLICATION_KEY,
  appSecret: process.env.OVH_APPLICATION_SECRET,
  consumerKey: process.env.OVH_CONSUMER_KEY
});

// ovhcloud domain-zone record create "${DNS_DOMAIN}" --field-type A --sub-domain "${SUB_DOMAIN}" --target "${VM_PUBLLIC_IP}" --ttl 3600

const new_record ={
  "fieldType": "A",
  "subDomain": process.argv[3],
  "target": process.argv[4],
  "ttl": 0
};

ovh.request('POST', `/domain/zone/${process.argv[2]}/record/`, new_record, function (err, zlist) {
  console.log(err || zlist);
});

