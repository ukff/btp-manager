#!/bin/bash
cd "$(dirname "$0")"

conditions=$(cat ../../controllers/conditions.go)

reasonsgo=$( echo $conditions | awk -v FS="(// gophers_reasons_section_start|// gophers_reasons_section_end)" '{print $2}')
reasons=$(echo $reasonsgo | grep -o '"[^"]*"')
echo "$reasons"

printf "===="

datago=$( echo $conditions | awk -v FS="(// gophers_metadata_section_start|// gophers_metadata_section_end)" '{print $2}')
cleandatago=$( echo $datago | awk -v FS="({|})" '{print $2}')
echo "$cleandatago"
 
printf "===="

operationsmd=$(cat ../../docs/operations.md)
table=$( echo "$operationsmd" | sed -n '/\/\/ gophers_table_start/,/\/\/ gophers_table_end/p')
echo "$table"