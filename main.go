package main

import (
	"bufio"
	"fmt"
	"glox/lox"
	"io/ioutil"
	"os"
)

/* additions :

   declare constant  e.g  const a = 1;
   loop break/continue
   string multiply by number ( a la python, e.g  "@" * 3 ,  3 * "@" = "@@@" )

   TODO:
   add switch statement
   debug "print "hello "+str(2);" runtime crash
*/

func main() {

	var do_repl bool
	var filename string

	fmt.Println("GLOX V0.01")
	vm := lox.NewVM()

	if len(os.Args) == 1 {
		usage()
	}
	for _, arg := range os.Args {
		switch arg {
		case "--debug":
			lox.DebugPrintCode = true
			lox.DebugTraceExecution = true
		case "--repl":
			do_repl = true
		default:
			filename = arg
		}
	}
	if do_repl {
		repl(vm)
	} else {
		if filename == "" {
			usage()
		}
		runFile(filename, vm)
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
				fmt.Println(result)
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
