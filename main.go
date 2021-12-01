package main

import (
	"bufio"
	"fmt"
	"glox/lox"
	"io/ioutil"
	"os"
)

/*


 */

var debugger = true

func main() {

	var do_repl bool
	var args = []string{}

	fmt.Println("GLOX:")
	vm := lox.NewVM()

	if debugger {
		lox.DebugPrintCode = true
		lox.DebugTraceExecution = true
		runFile("nod.lox", vm)
		os.Exit(0)
	}

	if len(os.Args) == 1 {
		usage()
	}
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--debug":
			lox.DebugPrintCode = true
			lox.DebugTraceExecution = true
		case "--repl":
			do_repl = true
		default:
			args = append(args, arg)
		}
	}

	if do_repl {
		repl(vm)
	} else {
		if len(args) == 0 {
			usage()
		}
		vm.SetArgs(args)
		runFile(args[0], vm)
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

func runFile(path string, vm *lox.VM) {

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Could not open file %s.", path)
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

func usage() {
	fmt.Println("Usage : glox [--debug][--repl] filename")
	os.Exit(1)
}
