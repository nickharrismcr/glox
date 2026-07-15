package main

import (
	"bufio"
	"fmt"
	"glox/src/compiler"
	"glox/src/core"
	dbg "glox/src/debug"
	"glox/src/vm"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"time"
)

// raylib/GLFW requires every call to happen on the OS thread that created the
// window. Without pinning the main goroutine, the Go scheduler is free to
// migrate it to a different OS thread after any blocking point (e.g. the
// worker-goroutine wg.Wait() in the Julia/Mandelbrot array builtins), which
// silently corrupts the GL command stream -- observed as intermittent
// flickering/corrupted pixels in graphics examples.
func init() {
	runtime.LockOSThread()
}

type Options struct {
	doRepl      bool
	printTokens bool
	cpuProfile  string
	memProfile  string
	args        []string
}

func main() {
	opts := parseArgs()

	if opts.doRepl {
		fmt.Println("GLOX:")
		vmInstance := vm.NewVM("repl", true)
		vmInstance.SetRepl(true)
		repl(vmInstance)
		return
	}

	if len(opts.args) == 0 {
		usage()
	}

	runFile(opts)
}

func parseArgs() *Options {
	if len(os.Args) == 1 {
		usage()
	}
	opts := &Options{args: []string{}}
	rawArgs := os.Args[1:]
	for i := 0; i < len(rawArgs); i++ {
		arg := rawArgs[i]
		if arg[0] == '-' {
			switch arg {
			case "--info":
				core.DebugTraceExecution = true
				core.LogLevel = core.INFO
			case "--debug", "-d":
				core.DebugPrintCode = true
				core.DebugTraceExecution = true
				core.LogLevel = core.TRACE
			case "--compile-only", "-c":
				core.DebugCompileOnly = true
			case "--globals", "-g":
				core.DebugShowGlobals = true
			case "--skip-builtins", "-s":
				core.DebugSkipBuiltins = true
			case "--repl":
				opts.doRepl = true
			case "--force-compile", "-f":
				core.ForceModuleCompile = true
			case "--print-tokens", "-p":
				opts.printTokens = true
			case "--instrument", "-i":
				core.DebugInstrument = true
			case "--no-peephole", "-n":
				core.DebugSkipPeephole = true
			case "--cpuprofile":
				i++
				if i >= len(rawArgs) {
					usage()
				}
				opts.cpuProfile = rawArgs[i]
			case "--memprofile":
				i++
				if i >= len(rawArgs) {
					usage()
				}
				opts.memProfile = rawArgs[i]
			default:
				usage()
			}
		} else {
			opts.args = append(opts.args, arg)
		}
	}
	return opts
}

// replInputComplete reports whether buffered REPL input forms a complete
// statement, i.e. all (){}[] are balanced and no string is left open. Strings,
// comments, and ${} interpolation never contribute stray/unbalanced brackets,
// so a simple net-depth count over the token stream is reliable.
func replInputComplete(src string) bool {
	s := compiler.NewScanner(src)
	depth := 0
	for _, t := range s.Tokens.Tokens {
		switch t.Tokentype {
		case compiler.TOKEN_LEFT_BRACE, compiler.TOKEN_LEFT_PAREN, compiler.TOKEN_LEFT_BRACKET:
			depth++
		case compiler.TOKEN_RIGHT_BRACE, compiler.TOKEN_RIGHT_PAREN, compiler.TOKEN_RIGHT_BRACKET:
			depth--
		case compiler.TOKEN_ERROR:
			if strings.Contains(t.Lexeme(), "Unterminated") {
				return false // open string/interpolation — keep reading
			}
		}
	}
	return depth <= 0
}

func repl(vmInstance *vm.VM) {
	inp := bufio.NewScanner(os.Stdin)
	var buf strings.Builder
	for {
		if buf.Len() == 0 {
			fmt.Print("> ")
		} else {
			fmt.Print("... ") // continuation prompt
		}
		if !inp.Scan() {
			return // EOF (Ctrl-Z / Ctrl-D)
		}
		line := inp.Text()

		if buf.Len() == 0 && len(line) == 0 {
			return // blank line at top level exits
		}
		if buf.Len() > 0 && len(line) == 0 {
			buf.Reset() // blank line while buffering cancels the pending entry
			continue
		}
		if buf.Len() > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(line)

		src := buf.String()
		if !replInputComplete(src) {
			continue // need more input
		}
		buf.Reset()

		status, result := vmInstance.Interpret(src, "__repl__")
		switch status {
		case vm.INTERPRET_OK:
			if result != "nil" {
				fmt.Println(result)
			}
		case vm.INTERPRET_RUNTIME_ERROR:
			fmt.Println(vmInstance.ErrorMsg)
			vmInstance.PrintStackTrace()
			// compile errors are already reported by the compiler as they occur
		}
	}
}

// warnIfNoDebugHook warns on stderr when a debug/instrument flag is used
// against a fast build that has the hot-loop debug hook compiled out (see
// core.HotLoopDebugHookCompiled and bin/build_debug.sh). Non-fatal: the
// script still runs, just without the requested output.
func warnIfNoDebugHook(flag, consequence string) {
	if core.HotLoopDebugHookCompiled {
		return
	}
	fmt.Fprintf(os.Stderr, "warning: this is a fast (non-debug) build -- the per-instruction hook that %s needs is compiled out, so %s. Rebuild with bin/build_debug.sh for a debug-capable binary.\n", flag, consequence)
}

func runFile(opts *Options) {
	args := opts.args

	// os.Exit skips deferred calls, so CPU profiling must be stopped and
	// flushed explicitly on every exit path once started.
	exit := func(code int) {
		if opts.cpuProfile != "" {
			pprof.StopCPUProfile()
		}
		os.Exit(code)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
			exit(1)
		}
	}()

	path := args[0]
	bytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Could not open file %s : %s", path, err)
		os.Exit(1)
	}
	source := string(bytes)

	if opts.cpuProfile != "" {
		f, err := os.Create(opts.cpuProfile)
		if err != nil {
			fmt.Printf("Could not create CPU profile %s : %s", opts.cpuProfile, err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Printf("Could not start CPU profile: %s", err)
			os.Exit(1)
		}
	}

	if opts.printTokens {
		compiler.PrintTokens(source)
		exit(0)
	}

	defineBuiltins := !core.DebugSkipBuiltins
	vmInstance := vm.NewVM(path, defineBuiltins)
	vmInstance.SetArgs(args)

	if core.DebugTraceExecution {
		warnIfNoDebugHook("--debug/--info", "trace output will be empty")
		vmInstance.DebugHook = dbg.TraceHook
	}
	if core.DebugInstrument {
		warnIfNoDebugHook("--instrument", "instruction counts will be zero")
		vmInstance.DebugHook = dbg.InstrumentHook
	}
	status, result := vmInstance.Interpret(source, "__main__")
	if status == vm.INTERPRET_COMPILE_ERROR {
		exit(65)
	}
	if status == vm.INTERPRET_RUNTIME_ERROR {
		fmt.Println(vmInstance.ErrorMsg)
		vmInstance.PrintStackTrace()
		exit(70)
	}

	if opts.cpuProfile != "" {
		pprof.StopCPUProfile()
	}
	if opts.memProfile != "" {
		f, err := os.Create(opts.memProfile)
		if err != nil {
			fmt.Printf("Could not create memory profile %s : %s", opts.memProfile, err)
			os.Exit(1)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Printf("Could not write memory profile: %s", err)
			os.Exit(1)
		}
	}

	fmt.Println(result)
	if core.DebugInstrument {
		endTime := time.Now()
		runtime := endTime.Sub(vmInstance.Starttime)
		fmt.Printf("Execution time: %s\n", runtime)
		fmt.Printf("Executed %d instructions\n", dbg.InstructionCount)
		fmt.Printf("Average instructions per second: %.2f\n", float64(dbg.InstructionCount)/runtime.Seconds())
		dbg.InstructionCount = 0 // Reset for next run
	}
}

func usage() {
	fmt.Println(`Usage: glox [options] filename

Options:
  --debug, -d           Enable debug mode (trace execution, print code)
  --info                Set log level to INFO and enable execution tracing
  --compile-only, -c    Compile only, do not execute
  --globals, -g         Show global variables in debug output
  --skip-builtins, -s   Do not define built-in functions
  --repl                Start interactive REPL
  --force-compile, -f   Force module recompilation
  --print-tokens, -p    Print tokens and exit
  --instrument, -i      Enable instruction counting and timing
  --no-peephole, -n     Skip the peephole optimiser
  --cpuprofile <file>   Write a CPU profile to <file>
  --memprofile <file>   Write a heap profile to <file> after execution`)
	os.Exit(1)
}
