package main

import "fmt"

func (vm *VM) compile(source string) {

	s := NewScanner(source)
	line := -1
	for {
		token := s.scanToken()
		if token.line != line {
			fmt.Printf("%4d ", token.line)
			line = token.line
		} else {
			fmt.Printf("   | ")
		}

		fmt.Printf("%-20s '%s'\n", token_names[token.tokentype], (*token.source)[token.start:token.start+token.length])
		if token.tokentype == TOKEN_EOF {
			break
		}
	}

}
