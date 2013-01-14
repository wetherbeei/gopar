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
	node     *ast.Node // CallExpr, RangeStmt, ForStmt
	target   *ast.Node // FuncDecl or FuncLit. Valid if node == ast.CallExpr
}

type Visitor struct {
	depth int
}

func (v Visitor) Visit(node ast.Node) (w ast.Visitor) {
	if node != nil {
		fmt.Println(strings.Repeat(". ", v.depth), reflect.TypeOf(node), node)
	}
	return Visitor{depth: v.depth + 1}
}

func buildCallGraph(project *Project) (root *Block) {
	mainFn := project.get("main").Lookup("main").Decl.(*ast.FuncDecl)
	var v Visitor
	ast.Walk(v, mainFn)

	// TODO: Recursively fill in imports, remember Go doesn't allow circular
	// dependencies
	return
}
