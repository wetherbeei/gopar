// Analyze the data dependencies between "basic blocks" (for loops + functions)
// Any dependencies within blocks don't matter, only those that cross the
// boundaries of loops and function calls. Each basic block records:
// - New identifiers defined
// - Identifiers read
// - Identifiers written
// External reads and writes can be calculated from that list. Dependencies
// from sub-blocks are carried upwards/outwards.
//
// Array accesses a[i+1]??
package main

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"
)

type DependencyPass struct {
	BasePass
}

type BasicBlock struct {
	parent   *BasicBlock
	children []*BasicBlock
	defines  []*ast.Ident
	reads    []*ast.Ident
	writes   []*ast.Ident
}

type DependencyPassVisitor struct {
	depth int
	cur   *BasicBlock
	node  ast.Node // the current node being visited
	pass  *DependencyPass
}

func (v *DependencyPassVisitor) Print(args ...interface{}) {
	depth := v.depth
	if v.node == nil {
		depth = depth - 1
	}
	args = append(args, 0)
	copy(args[1:], args[0:])
	args[0] = 1
	args[0] = strings.Repeat("  ", depth)
	fmt.Println(args...)
}

func (v DependencyPassVisitor) Visit(node ast.Node) (w ast.Visitor) {
	block := v.cur
	v.depth = v.depth + 1
	v.node = node
	if node == nil {
		// post-order actions (all sub-nodes have been visited)
		//
		// Merge the sub-node read/write list with the current read/write list, and
		// add the block to the current block's child list
		v.Print("done", reflect.TypeOf(v.node))
		return v
	}
	v.Print("start", reflect.TypeOf(v.node), v.node)
	// Return a new visitor for new BasicBlocks, else return the current visitor
	switch t := node.(type) {
	case *ast.ForStmt, *ast.RangeStmt, *ast.FuncDecl:
		v.Print("+ new block", t)
		newBlock := &BasicBlock{parent: block}
		block.children = append(block.children, newBlock)
		v.cur = newBlock
	case *ast.CallExpr:

		// tag this node with a reference to the surrounding BasicBlock so we can
		// fill in additional reads/writes once we resolve all functions
		v.Print("+ tagged for later pass")
		v.pass.SetResult(node, block)
	}

	// Analyze all AST nodes for reads, writes and defines
	switch t := node.(type) {
	case *ast.AssignStmt:
		v.Print("- Writes", t.Lhs)
		v.Print("- Reads", t.Rhs)
	}
	return v
}

func NewDependencyPass() *DependencyPass {
	return &DependencyPass{
		BasePass: NewBasePass(),
	}
}

func (pass *DependencyPass) GetPassType() PassType {
	return DependencyPassType
}

func (pass *DependencyPass) GetPassMode() PassMode {
	return FunctionPass
}

func (pass *DependencyPass) GetDependencies() []PassType {
	return []PassType{ExternalFunctionPassType}
}

func (pass *DependencyPass) RunFunctionPass(node ast.Node, c *Compiler) (modified bool, err error) {
	fnResult := c.GetPassResult(ExternalFunctionPassType, node).([]string)
	if len(fnResult) > 0 {
		fmt.Println("Skipping func", node)
		return
	}
	v := DependencyPassVisitor{cur: &BasicBlock{}, pass: pass}
	ast.Walk(v, node)
	return
}
