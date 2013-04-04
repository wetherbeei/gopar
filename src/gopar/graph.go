package main

import (
	"fmt"
	"go/ast"
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
		fmt.Printf("%s start %T %+v\n", strings.Repeat(sep, v.depth), node, node)
	} else {
		fmt.Printf("%s done %T\n", strings.Repeat("  ", v.depth-1), v.n)
	}
	return Visitor{depth: v.depth + 1, n: node}
}

func showGraph(project *Project, pkg string) {
	//ast.Print(project.fset, project.get("main").file)
	mainFile := project.get(pkg).file
	v := Visitor{depth: 0}
	ast.Walk(v, mainFile)

	return
}
