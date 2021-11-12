package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {

	fmt.Println("GLOX V0.01")
	vm := NewVM()
	vm.test()

	if len(os.Args) == 1 {
		repl(vm)
	} else if len(os.Args) == 2 {
		runFile(os.Args[2], vm)
	}
}

func repl(vm *VM) {

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
	if err == nil {
		fmt.Println("Could not open file.")
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
