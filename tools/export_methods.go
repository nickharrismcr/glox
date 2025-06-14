package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

// ExportMethodsAndRenameUsages capitalizes all method names and renames their uses.
func ExportMethodsAndRenameUsages(filename string, outFilename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Map of oldName -> newName for methods
	renamed := make(map[string]string)

	// First pass: rename method declarations
	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Name == nil {
			return true
		}
		name := fn.Name.Name
		if len(name) > 0 && !ast.IsExported(name) {
			newName := strings.Title(name)
			renamed[name] = newName
			fn.Name.Name = newName
		}
		return true
	})

	// Second pass: rename all usages
	ast.Inspect(node, func(n ast.Node) bool {
		// Rename selector expressions: obj.method -> obj.Method
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if newName, ok := renamed[sel.Sel.Name]; ok {
				sel.Sel.Name = newName
			}
		}
		// Rename direct calls: method() -> Method()
		if ident, ok := n.(*ast.Ident); ok {
			if newName, ok := renamed[ident.Name]; ok {
				ident.Name = newName
			}
		}
		return true
	})

	out, err := os.Create(outFilename)
	if err != nil {
		return err
	}
	defer out.Close()
	return printer.Fprint(out, fset, node)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: export_methods <input.go> <output.go>")
		return
	}
	err := ExportMethodsAndRenameUsages(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Exported methods and renamed usages written to", os.Args[2])
	}
}
