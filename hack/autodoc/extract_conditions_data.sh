#!/bin/bash
cd "$(dirname "$0")"
conditions=$(cat ../../controllers/conditions.go)
constReasonsInGo=$( echo "$conditions" | sed -n '/\/\/ gophers_reasons_section_start/,/\/\/ gophers_reasons_section_end/p')
constReasonsClean=$(echo "$constReasonsInGo" | awk 'NR > 1 {print $1}' RS='(' FS=')' )
echo "$constReasonsClean"
printf "===="
metadataReasonsInGo=$( echo "$conditions" | sed -n '/\/\/ gophers_metadata_section_start/,/\/\/ gophers_metadata_section_end/p')
metadataReasonsClean=$( echo "$metadataReasonsInGo" | awk 'NR > 1 {print $1}' RS='{' FS='}')
echo "$metadataReasonsClean"
printf "===="
printf "\n"
operations=$(cat ../../docs/operations.md)
table=$( echo "$operations" | sed -n '/\/\/ gophers_table_start/,/\/\/ gophers_table_end/p')
echo "$table"