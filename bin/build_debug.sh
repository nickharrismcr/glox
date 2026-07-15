#!/usr/bin/env bash
# Builds a debug-capable glox binary (bin/debug_glox[.exe]) with the
# per-instruction debug hook compiled into the run() dispatch loop.
#
# The default build (bin/build.sh, `go build -o bin/glox main.go`) has that
# hook commented out: its mere presence costs ~25% on dispatch-bound code
# (see docs/performance-roadmap.md Step 1), so it's off by default and
# --debug/--info/--instrument print a warning instead of silently doing
# nothing (see warnIfNoDebugHook in main.go).
#
# This script temporarily uncomments the hook line in src/vm/vm.go and
# flips core.HotLoopDebugHookCompiled in src/core/config.go, builds, then
# restores both files -- the working tree is unaffected before and after a
# run, whether or not the build succeeds.
set -e

cd "$(dirname "$0")/.."

VM_GO="src/vm/vm.go"
CONFIG_GO="src/core/config.go"

HOOK_OFF_LINE=$'\t\t// if vm.DebugHook != nil { vm.DebugHook(vm, core.DebugEventOpcode, inst) }'
HOOK_ON_LINE=$'\t\tif vm.DebugHook != nil { vm.DebugHook(vm, core.DebugEventOpcode, inst) }'

off_lineno=$(grep -nF "$HOOK_OFF_LINE" "$VM_GO" | head -1 | cut -d: -f1)
if [ -z "$off_lineno" ]; then
    echo "error: expected commented-out debug hook line not found in $VM_GO." >&2
    echo "It may already be uncommented (a previous run of this script didn't clean up) or hand-edited. Aborting without changes." >&2
    exit 1
fi
if ! grep -q 'const HotLoopDebugHookCompiled = false' "$CONFIG_GO"; then
    echo "error: expected 'const HotLoopDebugHookCompiled = false' not found in $CONFIG_GO. Aborting without changes." >&2
    exit 1
fi

restore() {
    sed -i -b "${off_lineno}s|.*|${HOOK_OFF_LINE}|" "$VM_GO"
    sed -i -b 's/const HotLoopDebugHookCompiled = true/const HotLoopDebugHookCompiled = false/' "$CONFIG_GO"
}
trap restore EXIT

sed -i -b "${off_lineno}s|.*|${HOOK_ON_LINE}|" "$VM_GO"
sed -i -b 's/const HotLoopDebugHookCompiled = false/const HotLoopDebugHookCompiled = true/' "$CONFIG_GO"

go build -o bin/debug_glox main.go
cp bin/debug_glox bin/debug_glox.exe

echo "Built bin/debug_glox (and bin/debug_glox.exe) -- per-instruction debug hook compiled in."
echo "Source restored to the default fast-build state (hook commented out)."
