package core

import "fmt"

func Log(level LoggingLevelType, s string) {
	if DebugTraceExecution && level >= LogLevel {
		fmt.Println(s)
	}
}

func LogFmt(level LoggingLevelType, format string, args ...interface{}) {
	if DebugTraceExecution && level >= LogLevel {
		fmt.Printf(format, args...)
	}
}
