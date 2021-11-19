package main

import (
	"lox"
	"testing"
)

type Test struct {
	input, output string
}
type Tests []Test

func TestVMExpression(t *testing.T) {

	vm := lox.NewVM()

	tests := Tests{
		{"1+2+3+4+5", "15.000000"},
		{"1+(2*(3+(4*-5)))", "-33.000000"},
		{"1+(2*3)+(4*5)", "27.000000"},
		{"1+(2/(3+(4/-5)))", "1.909091"},
		{"2>1", "true"},
		{"2>=1", "true"},
		{"2<1", "false"},
		{"2<=1", "false"},
		{"!true", "false"},
		{"!false", "true"},
		{"1==1", "true"},
		{"true!=false", "true"},
		{"false!=1", "true"},
		{"nil==nil", "true"},
		{"nil==false", "false"},
		{"nil==\"hello\"", "false"},
		{"\"hello\"==\"hello\"", "true"},
		{"\"hello\"!=\"hello\"", "false"},
		{"\"hello\"+\"hello\"", "hellohello"},
	}

	for i, test := range tests {
		status, res := vm.interpret(test.input)
		if status != lox.INTERPRET_OK || res != test.output {
			t.Errorf("Test %d failed : %s != %s ", i, test.output, res)
		}
	}
}

func TestVMSyntaxError(t *testing.T) {

	vm := lox.NewVM()

	status, _ := vm.interpret("//0ax")
	if status != lox.INTERPRET_COMPILE_ERROR {
		t.Error("Test failed : compile error expected ")
	}
	status, _ = vm.interpret("x9iu-")
	if status != lox.INTERPRET_COMPILE_ERROR {
		t.Error("Test failed : compile error expected ")
	}

}

func TestVMRuntimeError(t *testing.T) {

	vm := lox.NewVM()

	status, _ := vm.interpret("false+1")
	if status != lox.INTERPRET_RUNTIME_ERROR {
		t.Error("Test failed : runtime error expected ")
	}
	status, _ = vm.interpret("\"hello\"+1")
	if status != lox.INTERPRET_RUNTIME_ERROR {
		t.Error("Test failed : runtime error expected ")
	}

}
