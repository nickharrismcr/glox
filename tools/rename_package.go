package main

import (
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

func renamePackage(filename, outFilename, newPkg string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.PackageClauseOnly)
	if err != nil {
		return err
	}

	node.Name.Name = newPkg

	// Re-parse the full file to preserve the rest of the code
	fullNode, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	fullNode.Name.Name = newPkg

	out, err := os.Create(outFilename)
	if err != nil {
		return err
	}
	defer out.Close()
	return printer.Fprint(out, fset, fullNode)
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: rename_package <input.go> <output.go> <new_package>")
		return
	}
	err := renamePackage(os.Args[1], os.Args[2], os.Args[3])
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Renamed package written to", os.Args[2])
	}
}
