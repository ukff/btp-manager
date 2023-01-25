#!/bin/bash

helm template $1 module-chart --output-dir rendered --values public-overrides.yaml
mv rendered/sap-btp-operator/templates/ ../module-resources
rm -r rendered/