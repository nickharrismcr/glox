package main

import (
	"bufio"
	"fmt"
	"glox/lox"
	"io/ioutil"
	"os"
)

/* additions to vanilla lox :

   declare constant  e.g  const a = 1;
   modulus operator %
   loop break/continue
   string multiply by number ( a la python, e.g  "@" * 3 ,  3 * "@" = "@@@" )
   native funcs :  str(number)   len(string)    sin(x)  cos(x)
   lists  ( list literal initialisers (a=[1,2,3];),  indexing (a[x]),  slicing (a[x:y]))
   string slices   ( a = "abcd"; a[0]=="a", a[:2]=="ab", etc)

   TODO:
   list del  (del(list, index))
   add switch statement
   integer number type

   maps

*/

func main() {

	var do_repl bool
	var filename string

	fmt.Println("GLOX:")
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
