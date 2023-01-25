#!/bin/bash

curl -J -O -L https://raw.githubusercontent.com/ukff/btp-manager/charts-handling/module-overrides/public-overrides.yaml
helm template $1 module-chart --output-dir rendered --values public-overrides.yaml
mv rendered/sap-btp-operator/templates/ module-resources
rm -r rendered/