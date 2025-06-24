package main

import (
	"bufio"
	"fmt"
	lox "glox/src"
	"glox/src/compiler"
	"glox/src/core"
	dbg "glox/src/debug"
	"os"
	"runtime/debug"
)

type Options struct {
	doRepl      bool
	printTokens bool
	args        []string
}

func main() {

	// // // Start CPU profiling
	// f, err := os.Create("cpu.pprof")
	// if err != nil {
	// 	log.Fatal("could not create CPU profile: ", err)
	// }
	// if err := pprof.StartCPUProfile(f); err != nil {
	// 	log.Fatal("could not start CPU profile: ", err)
	// }
	// defer func() {
	// 	pprof.StopCPUProfile()
	// 	f.Close()
	// }()

	opts := &Options{}

	handleArgs(opts)

	if opts.doRepl {
		fmt.Println("GLOX:")
		vm := lox.NewVM("repl", true)
		repl(vm)
	} else {
		if len(opts.args) == 0 {
			usage()
		}
		runFile(opts)
	}
}

func repl(vm *lox.VM) {

	inp := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("> ")
		for inp.Scan() {
			s := inp.Text()
			if len(s) == 0 {
				return
			}
			status, result := vm.Interpret(s, "__repl__")
			if status == lox.INTERPRET_OK {
				if result != "nil" {
					fmt.Println(result)
				}
			}
			break
		}
	}
}

func runFile(opts *Options) {

	args := opts.args
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
			os.Exit(1)
		}
	}()

	path := args[0]
	bytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Could not open file %s : %s", path, err)
		os.Exit(1)
	}
	source := string(bytes)

	if opts.printTokens {
		compiler.PrintTokens(source)
		os.Exit(0)
	}

	defineBuiltins := !core.DebugSkipBuiltins
	vm := lox.NewVM(path, defineBuiltins)
	vm.SetArgs(args)
	if core.DebugTraceExecution {
		vm.DebugHook = dbg.TraceHook
	}
	status, result := vm.Interpret(source, "__main__")
	if status == lox.INTERPRET_COMPILE_ERROR {
		os.Exit(65)
	}
	if status == lox.INTERPRET_RUNTIME_ERROR {
		fmt.Println(vm.ErrorMsg)
		vm.PrintStackTrace()
		os.Exit(70)
	}
	fmt.Println(result)
}

func handleArgs(opts *Options) {

	if len(os.Args) == 1 {
		usage()
	}

	opts.args = []string{}

	for _, arg := range os.Args[1:] {
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
			default:
				usage()
			}
		} else {
			opts.args = append(opts.args, arg)
		}
	}
}

func usage() {
	fmt.Println("Usage : glox [--debug][--globals][--skip-builtins][--repl] filename")
	os.Exit(1)
}
