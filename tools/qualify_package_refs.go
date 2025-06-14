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

// getExportedStructsMethodsFields scans a package dir and returns exported structs, methods, and fields.
func getExportedStructsMethodsFields(pkgDir string) (map[string]struct{}, map[string]struct{}, map[string]map[string]struct{}, error) {
	structs := make(map[string]struct{})
	methods := make(map[string]struct{})
	fields := make(map[string]map[string]struct{}) // StructName -> set of FieldNames

	err := filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		for _, decl := range node.Decls {
			// Structs and fields
			if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.TYPE {
				for _, spec := range gd.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if st, ok := ts.Type.(*ast.StructType); ok && ast.IsExported(ts.Name.Name) {
							structs[ts.Name.Name] = struct{}{}
							if fields[ts.Name.Name] == nil {
								fields[ts.Name.Name] = make(map[string]struct{})
							}
							for _, field := range st.Fields.List {
								for _, name := range field.Names {
									if ast.IsExported(name.Name) {
										fields[ts.Name.Name][name.Name] = struct{}{}
									}
								}
							}
						}
					}
				}
			}
			// Methods
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv != nil && ast.IsExported(fn.Name.Name) {
				methods[fn.Name.Name] = struct{}{}
			}
		}
		return nil
	})
	return structs, methods, fields, err
}

// qualifyReferences rewrites identifiers in a file to use the package qualifier.

func qualifyReferences(filename, outFilename, pkg string, structs, methods map[string]struct{}, fields map[string]map[string]struct{}) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Add import if missing
	needImport := true
	for _, imp := range node.Imports {
		if strings.Trim(imp.Path.Value, `"`) == pkg {
			needImport = false
			break
		}
	}

	// Helper: recursively walk AST with parent tracking
	var walk func(ast.Node, ast.Node)
	walk = func(n ast.Node, parent ast.Node) {
		if n == nil {
			return
		}
		switch ident := n.(type) {
		case *ast.Ident:
			// Only qualify if not part of a SelectorExpr (i.e., not foo.Bar)
			if _, ok := structs[ident.Name]; ok {
				if _, isSel := parent.(*ast.SelectorExpr); !isSel {
					ident.Name = pkg + "." + ident.Name
				}
			}
			if _, ok := methods[ident.Name]; ok {
				if _, isSel := parent.(*ast.SelectorExpr); !isSel {
					ident.Name = pkg + "." + ident.Name
				}
			}
		case *ast.SelectorExpr:
			// Qualify struct field usages: obj.Field
			if id, ok := ident.X.(*ast.Ident); ok {
				structName := id.Name
				fieldName := ident.Sel.Name
				if fieldSet, ok := fields[structName]; ok {
					if _, ok := fieldSet[fieldName]; ok {
						// Qualify struct name if not already qualified
						if !strings.Contains(structName, ".") {
							id.Name = pkg + "." + structName
						}
					}
				}
			}
		}
		// Recurse into children
		ast.Inspect(n, func(child ast.Node) bool {
			if child != nil && child != n {
				walk(child, n)
			}
			return true
		})
	}

	walk(node, nil)

	if needImport {
		newImport := &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf(`"%s"`, pkg),
			},
		}
		node.Imports = append(node.Imports, newImport)
	}

	out, err := os.Create(outFilename)
	if err != nil {
		return err
	}
	defer out.Close()
	return printer.Fprint(out, fset, node)
}
func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: qualify_package_refs <package_dir> <input.go> <output.go>")
		return
	}
	pkgDir := os.Args[1]
	input := os.Args[2]
	output := os.Args[3]

	// Use the last element of the package path as the package name
	pkgName := filepath.Base(pkgDir)

	structs, methods, fields, err := getExportedStructsMethodsFields(pkgDir)
	if err != nil {
		fmt.Println("Error scanning package:", err)
		return
	}

	err = qualifyReferences(input, output, pkgName, structs, methods, fields)
	if err != nil {
		fmt.Println("Error qualifying references:", err)
		return
	}
	fmt.Println("Qualified references written to", output)
}
