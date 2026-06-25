#! /usr/bin/env bash
. ./setenv
cd tests
python -m pytest new_tests/ $*
