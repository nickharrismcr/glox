#!/usr/bin/env bash
# Run each lox_examples/*.lox for 3 seconds, then kill it.
# PASS: process was killed by timeout (or exited cleanly) with no exception/panic output.
# FAIL: non-timeout crash (exit 65/70/1) OR stdout/stderr contains exception or panic text.

. "$(dirname "$0")/../setenv"
cp bin/glox bin/glox.exe 2>/dev/null || true

PASS=0
FAIL=0
ERRORS=()

contains_exception() {
    # Runtime errors: main.go prints ErrorMsg to stdout (exit 70)
    # Recovered panics: main.go prints recover() value to stdout (exit 1)
    # Go panics:        runtime prints goroutine trace to stderr
    grep -qE \
        "Uncaught exception:|panic:|unknown tag|goroutine [0-9]+|Runtime error:" \
        "$1" "$2" 2>/dev/null
}

for f in lox_examples/*.lox; do
    name=$(basename "$f")
    tmpout=$(mktemp)
    tmperr=$(mktemp)

    exit_code=0
    timeout 3 glox "$f" >"$tmpout" 2>"$tmperr" || exit_code=$?

    if [ "$exit_code" -eq 124 ] || [ "$exit_code" -eq 0 ]; then
        if contains_exception "$tmpout" "$tmperr"; then
            msg=$(grep -hE "Uncaught exception:|panic:|unknown tag|goroutine [0-9]+|Runtime error:" \
                      "$tmpout" "$tmperr" 2>/dev/null | head -1)
            echo "FAIL: $name"
            echo "      $msg"
            FAIL=$((FAIL + 1))
            ERRORS+=("$name")
        else
            echo "PASS: $name"
            PASS=$((PASS + 1))
        fi
    else
        first_line=$(head -1 "$tmpout" 2>/dev/null)
        echo "FAIL: $name (exit $exit_code${first_line:+: $first_line})"
        FAIL=$((FAIL + 1))
        ERRORS+=("$name")
    fi

    rm -f "$tmpout" "$tmperr"
done

echo ""
echo "Results: $PASS passed, $FAIL failed"

if [ $FAIL -gt 0 ]; then
    echo "Failed: ${ERRORS[*]}"
    exit 1
fi
