#!/bin/bash
cd "$(dirname "$0")"

trap 'catch' ERR
catch() {
    echo "error"
    exit 1
}

readonly CHART_PATH="../../module-chart/chart"
readonly CHART_OVERRIDES_PATH="../../module-chart/overrides.yaml"
readonly EXISTING_RESOURCES_PATH="../../module-resources"
readonly EXISTING_RESOURCES_DELETE_PATH="../../module-resources/delete"
readonly EXISTING_RESOURCES_APPLY_PATH="../../module-resources/apply"
readonly HELM_OUTPUT_PATH="rendered"
readonly NEW_RESOURCES_PATH="rendered/sap-btp-operator/templates"

tag=$1

helm template $tag $CHART_PATH --output-dir $HELM_OUTPUT_PATH --values $CHART_OVERRIDES_PATH --namespace "kyma-system"

trap 'rm -rf -- "temp"' EXIT
runActionForEachYaml() {
  local directory=${1}
  local action=${2}

  if [ "$(ls -A $directory)" ]; then    
    for combinedYaml in $directory/*
    do
        mkdir 'temp' && cd 'temp'
        yq -s '"file_" + $index' "../$combinedYaml"
        for singleYaml in *
        do
          $action $singleYaml
        done
        cd .. && rm -r 'temp'
    done
  else
    echo "$directory is Empty"
  fi
}

actionForNewResource() {
  local yaml=${1}
  incoming_resources+=("$(yq '.metadata.name' $yaml):$(yq '.kind' $yaml)")
}

actionForExistingResource() {
    local yaml=${1}
    if [[ ! "${incoming_resources[*]}" =~ "$(yq '.metadata.name' $yaml):$(yq '.kind' $yaml)" ]] ; then
        cat $yaml >> ../to-delete.yml
    fi
}

incoming_resources=()

runActionForEachYaml $NEW_RESOURCES_PATH actionForNewResource

touch to-delete.yml
runActionForEachYaml $EXISTING_RESOURCES_APPLY_PATH actionForExistingResource

rm -r $EXISTING_RESOURCES_PATH
mkdir $EXISTING_RESOURCES_PATH
mkdir $EXISTING_RESOURCES_APPLY_PATH
mkdir $EXISTING_RESOURCES_DELETE_PATH
mv $NEW_RESOURCES_PATH/* $EXISTING_RESOURCES_APPLY_PATH
mv to-delete.yml $EXISTING_RESOURCES_DELETE_PATH
rm -r $HELM_OUTPUT_PATH
