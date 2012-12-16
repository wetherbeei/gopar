package main

import (
	"go/ast"
)

type ParallelFunc struct {
	f             *ast.File
	decl          *ast.FuncDecl
	parallelLoops []*ast.RangeStmt
}

// FindParallelFuncs returns a list of references to functions to be
// parallelized. A parallelizable function is an anonymous goroutine function
// launch containing one or more parallelizable range statements.
func FindParallelFuncs(f *ast.File) (funcs []*ParallelFunc) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.GoStmt:

		}
		return true
	})
	return
}
