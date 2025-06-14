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

// ExportStructFieldsAndRenameUsages capitalizes all struct field names and renames their usages.
func ExportStructFieldsAndRenameUsages(filename string, outFilename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Map of oldName -> newName for struct fields
	renamed := make(map[string]string)

	// First pass: rename struct field declarations
	ast.Inspect(node, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return true
		}
		for _, field := range st.Fields.List {
			for _, name := range field.Names {
				if len(name.Name) > 0 && !ast.IsExported(name.Name) {
					newName := strings.Title(name.Name)
					renamed[name.Name] = newName
					name.Name = newName
				}
			}
		}
		return true
	})

	// Second pass: rename all usages (selector expressions)
	ast.Inspect(node, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if newName, ok := renamed[sel.Sel.Name]; ok {
				sel.Sel.Name = newName
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
		fmt.Println("Usage: export_struct_fields <input.go> <output.go>")
		return
	}
	err := ExportStructFieldsAndRenameUsages(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Exported struct fields and renamed usages written to", os.Args[2])
	}
}
