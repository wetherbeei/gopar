package main

import (
	"go/ast"
)

type Block struct {
	parent   *Block
	children []*Block
	node     *ast.Node     // CallExpr, RangeStmt, ForStmt
	target   *ast.FuncDecl // valid if node == ast.CallExpr
	next     *Block        // linked list of Blocks on this level
}

func buildCallGraph(project *Project) (root *Block) {
	// Generate a map of all functions
	//functions := make(map[string]*Block)

	// TODO: Recursively fill in imports, remember Go doesn't allow circular
	// dependencies
	return
}
