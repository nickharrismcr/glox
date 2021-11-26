package main

import (
	"bufio"
	"fmt"
	"glox/lox"
	"io/ioutil"
	"os"
)

/*
	todo https://craftinginterpreters.com/classes-and-instances.html#class-objects

	# GLOX

	**Bob Nystroms clox bytecode interpreter implemented in Go**
	Cop out : GC is handled by the Go runtime.

	**Additions to vanilla Lox:**

	immutable vars e.g  const a = 1;
	modulus operator %
	loop break/continue
	string multiply by number ( a la python, e.g  "@" * 3 ,  3 * "@" = "@@@" )
	native funcs :  str(value)    len(string|list)      sin(x)    cos(x)     append(list,value)

	lists :
		initialiser (a=[]; a=[1,2,3];)
		indexing ( b=a[x] )
		slicing ( b=a[x:y]; b=a[:y]; b=a[x:]; b=a[:] )
		adding ( list3=list1+list2 )
		appends ( native append(list,val) )

	string slices   ( a = "abcd"; b=a[0], b=a[:2], etc)
	renamed fun to func (!)

	TODO:

	Bob's classes chapter
	 -  allow class __str__ magic method to define str()/print output
	command line arguments (e.g sys.argv[])
	list index/slice assignment ( a[1]="a" or a[2:5] = [1,2,3] )
	list item del  (del a[b] or del a[b:c] - i.e assign nil )
	- should be doable once the class stuff is in.
*/

func main() {

	var do_repl bool
	var filename string

	fmt.Println("GLOX:")
	vm := lox.NewVM()
	//runFile("nod.lox", vm)
	//os.Exit(0)

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
