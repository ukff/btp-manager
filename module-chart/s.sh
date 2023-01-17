#! /bin/bash

notApplicable=''
 
yq overrides.yaml -o props | \
while read i
do
  echo $i
done