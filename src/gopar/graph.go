package main

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"
)

type Block struct {
	parent   *Block
	children []*Block
	node     ast.Node // CallExpr, ForStmt
}

type Visitor struct {
	depth int
	block *Block
	n     ast.Node
}

func (v Visitor) Visit(node ast.Node) (w ast.Visitor) {
	block := v.block
	sep := ". "
	if node != nil {
		switch t := node.(type) {
		case *ast.RangeStmt, *ast.ForStmt:
			sep = "> "
			block = &Block{parent: v.block, node: t}
			v.block.children = append(v.block.children, block)
		case *ast.CallExpr:
			// TOOD: follow function calls
			sep = "= "
		}
		fmt.Println(strings.Repeat(sep, v.depth), "start", reflect.TypeOf(node), node)
	} else {
		fmt.Println(strings.Repeat("  ", v.depth-1), "done", reflect.TypeOf(v.n))
	}
	return Visitor{block: block, depth: v.depth + 1, n: node}
}

func buildCallGraph(project *Project) (root *Block) {
	//ast.Print(project.fset, project.get("main").file)
	mainFn := project.get("main").Lookup("main").Decl.(*ast.FuncDecl)
	v := Visitor{depth: 0, block: &Block{node: nil}}
	ast.Walk(v, mainFn)

	// TODO: Recursively fill in imports, remember Go doesn't allow circular
	// dependencies
	return v.block
}
