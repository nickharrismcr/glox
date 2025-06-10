package main

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

func injectTraceToFuncs(f *ast.File) {
	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if ok && fn.Body != nil {
			traceCall := &ast.ExprStmt{
				X: &ast.CallExpr{
					Fun:  ast.NewIdent("Debug"),
					Args: []ast.Expr{ast.NewIdent(`"` + fn.Name.Name + `"`)},
				},
			}
			// Insert trace() at the beginning of the function body
			fn.Body.List = append([]ast.Stmt{traceCall}, fn.Body.List...)
		}
		return true
	})
}

func main() {
	if len(os.Args) < 2 {
		println("Usage: labeltrace <file.go>")
		return
	}
	filename := os.Args[1]
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	injectTraceToFuncs(node)

	out, err := os.Create("traced_" + filename)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	printer.Fprint(out, fset, node)
}
