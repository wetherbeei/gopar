package main

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"
)

type Visitor struct {
	depth int
	n     ast.Node
}

func (v Visitor) Visit(node ast.Node) (w ast.Visitor) {
	sep := ". "
	if node != nil {
		switch node.(type) {
		case *ast.RangeStmt, *ast.ForStmt:
			sep = "> "
		case *ast.CallExpr:
			// TOOD: follow function calls
			sep = "= "
		}
		fmt.Println(strings.Repeat(sep, v.depth), "start", reflect.TypeOf(node), node)
	} else {
		fmt.Println(strings.Repeat("  ", v.depth-1), "done", reflect.TypeOf(v.n))
	}
	return Visitor{depth: v.depth + 1, n: node}
}

func showGraph(project *Project) {
	//ast.Print(project.fset, project.get("main").file)
	mainFn := project.get("main").Lookup("main").Decl.(*ast.FuncDecl)
	v := Visitor{depth: 0}
	ast.Walk(v, mainFn)

	return
}
