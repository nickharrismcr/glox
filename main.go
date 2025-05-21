package main

import (
	"bufio"
	"fmt"
	"glox/lox"
	"io/ioutil"
	"os"
)

type Options struct {
	debugger bool
	doRepl   bool
	args     []string
}

func main() {

	fmt.Println("GLOX:")

	opts := &Options{
		debugger: false,
	}

	if opts.debugger {
		runFile([]string{"dbg.lox"})
		os.Exit(0)
	}

	handleArgs(opts)

	if opts.doRepl {
		vm := lox.NewVM("repl")
		repl(vm)
	} else {
		if len(opts.args) == 0 {
			usage()
		}
		runFile(opts.args)
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
			status, result := vm.Interpret(s)
			if status == lox.INTERPRET_OK {
				if result != "nil" {
					fmt.Println(result)
				}
			}
			break
		}
	}
}

func runFile(args []string) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			os.Exit(1)
		}
	}()

	path := args[0]
	vm := lox.NewVM(path)
	vm.SetArgs(args)

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Could not open file %s : %s", path, err)
		os.Exit(1)
	}
	status, result := vm.Interpret(string(bytes))
	if status == lox.INTERPRET_COMPILE_ERROR {
		os.Exit(65)
	}
	if status == lox.INTERPRET_RUNTIME_ERROR {
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
			case "--debug":
				lox.DebugPrintCode = true
				lox.DebugTraceExecution = true
				lox.DebugShowGlobals = false
			case "--repl":
				opts.doRepl = true
			default:
				usage()
			}
		} else {
			opts.args = append(opts.args, arg)
		}
	}
}

func usage() {
	fmt.Println("Usage : glox [--debug][--repl] filename")
	os.Exit(1)
}
