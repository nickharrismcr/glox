package main

import "fmt"

//TODO move to go test framework

func (vm *VM) test() {

	fmt.Println("Running tests...")

	vm.runTest(1, "1+2+3+4+5", "15.000000")
	vm.runTest(2, "1+(2*3)+(4*5)", "27.000000")
	vm.runTest(3, "1+(2*(3+(4*-5)))", "-33.000000")
	vm.runTest(4, "1+(2/(3+(4/-5)))", "1.909091")
	vm.expectCompileError(5, "1+(/2/(3+(4/-5))-)")
	vm.expectCompileError(6, "1+a2")

	fmt.Println("Tests completed")
}

func (vm *VM) runTest(count int, testExpr, expected string) {
	status, res := vm.interpret(testExpr)
	if status != INTERPRET_OK || res != expected {
		panic(fmt.Sprintf("Test %d failed : %s != %s ", count, expected, res))
	}
}

func (vm *VM) expectCompileError(count int, testExpr string) {
	status, _ := vm.interpret(testExpr)
	if status != INTERPRET_COMPILE_ERROR {
		panic(fmt.Sprintf("Test %d failed : expected compile error ", count))
	}
}
