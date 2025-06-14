#!/bin/bash
search="$1"
replace="$2"
find . -name "*.go" -exec perl -pi -e "s/$search/$replace/g" {} +