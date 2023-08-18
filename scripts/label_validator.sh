#!/usr/bin/env bash

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # must be set if you want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

GITHUB_ORG="ukff"

# From Github API Docs why using API for Issue.
#   You can use the REST API to create comments on issues and pull requests. Every pull request is an issue, but not every issue is a pull request.

# Event which trigger execution of script. Can be RELEASE or PR
TRIGGER_EVENT=$1 
PR_ID=${2:-NA}

function runOnRelease() {
  latest=$(curl -H "X-GitHub-Api-Version: 2022-11-28" \
                -sS "https://api.github.com/repos/$GITHUB_ORG/btp-manager/releases/latest" | 
                jq -r '.tag_name')

  notValidPrs=()
  while read -r commit; do

    pr_id=$(curl -L \
              -H "Accept: application/vnd.github+json" \
              -H "X-GitHub-Api-Version: 2022-11-28" \
              "https://api.github.com/search/issues?q=$commit+repo:$GITHUB_ORG/btp-manager+type:pr" |
              jq 'if (.items | length) == 1 then .items[0].number else empty end')

    if [[ -z $pr_id ]]; then
      echo "not found PR number for given commit $commit"
      continue
    fi 

    pr_labels=$(curl -sL \
                    -H "Accept: application/vnd.github+json" \
                    -H "X-GitHub-Api-Version: 2022-11-28" \
                    https://api.github.com/repos/$GITHUB_ORG/btp-manager/issues/${pr_id} | 
                    jq -r '.labels[] | objects | .name')

    kind_labels=$(grep -o kind <<< ${pr_labels[*]} | wc -l)
    if [[ $kind_labels -ne 1 ]]; then 
      notValidPrs+=("$pr_id")
    fi
    
  done <<< "$(git log "$latest"..HEAD --pretty=tformat:"%h")"

  if [ ${#notValidPrs[@]} -gt 0 ]; then
      echo "following PRs do not have any kind label"
      echo "${notValidPrs[@]}"
      exit 1
  fi

  echo "label validation OK"
  exit 0
} 

function runOnPr() {
  if [[ $PR_ID == "NA" ]]; then
    echo "PR ID not given"
    exit 1
  fi

  supported_labels=()

  help_message="**Add one of following labels** <br/><br/>"

  while IFS= read -r label; do
    parts=$(tr "#" " " <<< "$label")
    set $parts
    label_part=$1; help_message_part=$2
    help_message="${help_message} - $label_part -> $help_message_part <br/><br/>"
    supported_labels+=("$label_part")
  done <<< "$(yq eval '.changelog.categories.[].labels' ./.github/release.yml | grep "\- kind"| sed -e 's/- //g')"

  comments=$(curl -sL \
              -H "Accept: application/vnd.github+json" \
              -H "X-GitHub-Api-Version: 2022-11-28" \
              https://api.github.com/repos/$GITHUB_ORG/btp-manager/issues/${PR_ID}/comments |
              jq -r '.[] | objects | .body')

  if [[ ! " ${comments[*]} " =~ " ${help_message} " ]]; then

    payload=$(jq -n \
      --arg body "$help_message" \
      '{
        "body": $body,
      }') 

    http_code=$(curl -sL \
                -w "%{http_code}" --output /dev/null \
                -X POST \
                -H "Accept: application/vnd.github+json" \
                -H "Authorization: Bearer $GITHUB_TOKEN" \
                -H "X-GitHub-Api-Version: 2022-11-28" \
                https://api.github.com/repos/$GITHUB_ORG/btp-manager/issues/${PR_ID}/comments \
                -d "$payload")
    if [[ "$http_code" != "201" ]]; then
      echo "Unable to create comment with help text. $http_code"
      exit 1
    fi
  fi

  present_labels=$(curl -sL \
                    -H "Accept: application/vnd.github+json" \
                    -H "X-GitHub-Api-Version: 2022-11-28" \
                    https://api.github.com/repos/$GITHUB_ORG/btp-manager/issues/${PR_ID} | 
                    jq -r '.labels[] | objects | .name')

  count_of_required_labels=$(grep -o -w -F -c "${supported_labels[*]}" <<< "$present_labels")
  if [[ $count_of_required_labels -eq 1 ]]; then 
    echo "label validation OK"
    exit 0
  fi

  echo "error: only 1 of following labels must be added to each PR before merge but found $count_of_required_labels:"
  echo "${supported_labels[@]}"
  exit 1
}

case $TRIGGER_EVENT in
  "RELEASE")
    runOnRelease
    ;;
  "PR")
    runOnPr
    ;;
  *)
    echo "unsupported trigger event: $TRIGGER_EVENT"
    exit 1
    ;;
esac
