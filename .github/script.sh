#!/bin/bash

ARG1=${1:-local}

latest=$(curl \
            -H "Accept: application/vnd.github+json" \
            -H "X-GitHub-Api-Version: 2022-11-28" \
            https://api.github.com/repos/$SAP/$BTP_OPERATOR_REPO/releases/latest | jq -r '.tag_name') 

echo "TAG=${latest}" >> $GITHUB_ENV

        curl -L https://github.com/$SAP/$BTP_OPERATOR_REPO/releases/download/$TAG/sap-btp-operator-$TAG.tgz > charts.tgz
        tar zxvf charts.tgz 
        mv sap-btp-operator module-chart

        curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
        chmod 700 get_helm.sh
        ./get_helm.sh

        curl -J -O -L https://raw.githubusercontent.com/ukff/btp-manager/charts-handling/module-overrides/public-overrides.yaml
        helm template $TAG . --output-dir rendered --values public-overrides.yaml
        mv rendered/sap-btp-operator/templates/ ./module-resources
        rm -r rendered/
