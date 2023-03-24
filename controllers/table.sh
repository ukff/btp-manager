#!/bin/bash

value=`cat conditions.go`

reasonsgo=$( echo $value | awk -v FS="(// XXX|// XXX)" '{print $2}')
reasons=$(echo $reasonsgo | grep -o '"[^"]*"')
echo $reasons

datago=$( echo $value | awk -v FS="(// XXX2|// XXX2)" '{print $2}')
cleandatago=$( echo $datago | awk -v FS="({|})" '{print $2}')

IFS=$'\n' read -rd '' -a y <<<"$cleandatago"

for i in "${y[@]}"
do
   echo "$i"
   echo '\n'
done