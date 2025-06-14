#!/bin/bash
# e.g replace.sh 'foo(\d+)bar' 'baz$1qux' 

file=$3
search="$1"
replace="$2" 
perl -pi -e "s/$search/$replace/g" "$file"
 