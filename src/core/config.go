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
