#!/usr/bin/env node

const ovh = require('@ovhcloud/node-ovh')({
  appKey: process.env.OVH_APPLICATION_KEY,
  appSecret: process.env.OVH_APPLICATION_SECRET,
  consumerKey: process.env.OVH_CONSUMER_KEY
});



ovh.request('DELETE', `/domain/zone/jcm.ovh/record/${process.argv[2]}` , function (err, zlist) {
  console.log(err || zlist);
});

