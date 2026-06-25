#!/usr/bin/env bash
set -e
. "$(dirname "$0")/../setenv"
cp bin/glox bin/glox.exe
echo "=== glox ==="
python bin/time_lox.py bin/fibo_benchmark.lox
echo "=== python ==="
python bin/time_lox.py --python bin/fibo_benchmark.py
