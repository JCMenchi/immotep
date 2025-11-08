Ajout de règle firewall

``` bash
gcloud compute --project=tutogc firewall-rules create allow-ssh --direction=INGRESS --priority=1000 --network=network-instance --action=ALLOW --rules=tcp:22 --source-ranges=0.0.0.0/0
```

Run playbook:
     ansible-playbook -i inventory.ini  vm.yml
