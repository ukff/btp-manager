#!/bin/bash
latest=$(curl \
        -H "Accept: application/vnd.github+json" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        https://api.github.com/repos/SAP/sap-btp-service-operator/releases/latest | jq -r '.tag_name') 
echo $latest
curl -L https://github.com/SAP/sap-btp-service-operator/releases/download/$latest/sap-btp-operator-$latest.tgz > charts.tgz
tar zxvf charts.tgz 
rsync -a sap-btp-operator/ module-chart/
rm -r sap-btp-operator
echo $latest