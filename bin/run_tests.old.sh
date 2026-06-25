#! /usr/bin/env bash
. ./setenv
cd tests 
python test.py $*
