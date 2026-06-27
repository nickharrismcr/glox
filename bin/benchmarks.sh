#!/usr/bin/env bash
set -e
. "$(dirname "$0")/../setenv"
cp bin/glox bin/glox.exe

BENCH_DIR="$(dirname "$0")/../benchmarks"
RUNS=${1:-3}

extract_avg() {
    grep "Average:" | sed 's/Average: \([0-9.]*\) seconds/\1/'
}

printf "\n%-20s %12s %12s %8s\n" "benchmark" "glox" "python" "ratio"
printf "%-20s %12s %12s %8s\n" "---------" "----" "------" "-----"

for lox_src in "$BENCH_DIR"/lox/*.lox; do
    name=$(basename "$lox_src" .lox)
    py_src="$BENCH_DIR/python/${name}.py"

    glox_avg=$(python bin/time_lox.py "$lox_src" --runs "$RUNS" 2>&1 | extract_avg)
    py_avg=$(python bin/time_lox.py --python "$py_src" --runs "$RUNS" 2>&1 | extract_avg)

    if [ -n "$glox_avg" ] && [ -n "$py_avg" ]; then
        ratio=$(python -c "print(f'{float(\"$glox_avg\")/float(\"$py_avg\"):.1f}x')")
    else
        ratio="err"
    fi

    printf "%-20s %11ss %11ss %8s\n" "$name" "$glox_avg" "$py_avg" "$ratio"
done
