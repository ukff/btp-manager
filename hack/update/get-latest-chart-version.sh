#!/bin/bash
cd "$(dirname "$0")"

trap 'catch' ERR
catch() {
    echo "error"
    exit 1
}

latest=$(curl \
        -H "Accept: application/vnd.github+json" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        https://api.github.com/repos/SAP/sap-btp-service-operator/releases/latest | jq -r '.tag_name')
echo $latest