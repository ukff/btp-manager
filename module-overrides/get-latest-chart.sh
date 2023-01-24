#!/bin/bash
        latest=$(curl \
                  -H "Accept: application/vnd.github+json" \
                  -H "X-GitHub-Api-Version: 2022-11-28" \
                  https://api.github.com/repos/$SAP/$BTP_OPERATOR_REPO/releases/latest | jq -r '.tag_name') 
        curl -L https://github.com/$SAP/$BTP_OPERATOR_REPO/releases/download/$TAG/sap-btp-operator-$TAG.tgz > charts.tgz
        tar zxvf charts.tgz 
        mv sap-btp-operator module-chart
        echo $latest