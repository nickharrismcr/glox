package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

/* additions :

   declare a variable as constant  e.g  const a = 1;

   TODO:
   add loop break/continue
    - don't create new opcodes?  vm doesn't know if we are in a loop, just follows jumps
	- compiler should insert necessary jumps
	- "Make sure to think about scope. What should happen to local variables declared inside the body
	   of the loop or in blocks nested inside the loop when a continue is executed?"
   add switch statement
*/

func main() {

	fmt.Println("GLOX V0.01")
	vm := NewVM()
	runFile("testloop.lox", vm)
	return

	if len(os.Args) == 1 {
		repl(vm)
	} else if len(os.Args) == 2 {
		runFile(os.Args[1], vm)
	}
}

func repl(vm *VM) {

	/* 	code := `
	var a = 1;
	{
		var b = 2;
		b= 3;
		print b;
	}`
		vm.interpret(code)
		return */

	inp := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("> ")
		for inp.Scan() {
			s := inp.Text()
			if len(s) == 0 {
				return
			}
			status, result := vm.interpret(s)
			if status == INTERPRET_OK {
				fmt.Println(result)
			}
			break
		}

	}

}

func runFile(path string, vm *VM) {

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Could not open file %s.", path)
		os.Exit(1)
	}
	status, result := vm.interpret(string(bytes))
	if status == INTERPRET_COMPILE_ERROR {
		os.Exit(65)
	}
	if status == INTERPRET_RUNTIME_ERROR {
		os.Exit(70)
	}
	fmt.Println(result)
}
