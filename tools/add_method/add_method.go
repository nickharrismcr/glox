package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := "./" // or your package path
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if !strings.HasPrefix(path, "obj_") {
			return nil // Skip files that don't start with "obj_"
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		// Collect all struct type names
		structTypes := make(map[string]bool)
		for _, decl := range node.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, ok := ts.Type.(*ast.StructType); ok {
					structTypes[ts.Name.Name] = false // not yet seen IsBuiltIn
				}
			}
		}

		// Mark types that already have IsBuiltIn
		for _, decl := range node.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Name.Name != "IsBuiltIn" {
				continue
			}
			if len(fn.Recv.List) > 0 {
				if starExpr, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
					if ident, ok := starExpr.X.(*ast.Ident); ok {
						structTypes[ident.Name] = true
					}
				}
			}
		}

		// Add IsBuiltIn stub for types that don't have it
		for typeName, hasMethod := range structTypes {
			if hasMethod {
				continue
			}
			// func (t *TypeName) IsBuiltIn() bool { return false }
			recv := &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{ast.NewIdent("t")},
						Type:  &ast.StarExpr{X: ast.NewIdent(typeName)},
					},
				},
			}
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("IsBuiltIn"),
				Recv: recv,
				Type: &ast.FuncType{
					Params:  &ast.FieldList{},
					Results: &ast.FieldList{List: []*ast.Field{{Type: ast.NewIdent("bool")}}},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent("false")},
						},
					},
				},
			}
			node.Decls = append(node.Decls, fn)
		}

		// Write back to file
		out, err := os.Create(path)
		if err != nil {
			return err
		}
		defer out.Close()
		return printer.Fprint(out, fset, node)
	})
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("IsBuiltIn stubs added where needed.")
	}
}
