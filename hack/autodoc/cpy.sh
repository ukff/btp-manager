#!/bin/bash
cd "$(dirname "$0")"

operationsmd=$(cat ../../docs/operations.md)
table=$( echo "$operationsmd" | sed -n '/\/\/ gophers_table_start/,/\/\/ gophers_table_end/p')
echo "$table"