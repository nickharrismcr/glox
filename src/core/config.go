package core

type LoggingLevelType int

const (
	TRACE LoggingLevelType = iota
	DEBUG
	INFO
	WARN
	ERROR
)

var DebugSuppress = false
var DebugTraceExecution = false
var DebugPrintCode = false
var DebugShowGlobals = false
var DebugSkipBuiltins = false
var DebugCompileOnly = false
var DebugInstrument = false
var DebugSkipPeephole = false
var LogLevel = INFO

var ForceModuleCompile = false

// HotLoopDebugHookCompiled reports whether the per-instruction debug-hook
// call in vm.go's run() dispatch loop is compiled in. False in the default
// (fast) build; bin/build_debug.sh flips this to true in lockstep with
// uncommenting the hook call itself -- see the comment there and
// docs/performance-roadmap.md Step 1. main.go uses this to warn when
// --debug/--info/--instrument are requested on a build that cannot honour
// them.
const HotLoopDebugHookCompiled = false
