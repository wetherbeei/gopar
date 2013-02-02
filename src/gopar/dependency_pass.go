// Analyze the data dependencies between "basic blocks" (for loops + functions)
// Any dependencies within blocks don't matter, only those that cross the
// boundaries of loops and function calls. Each basic block records:
// - New identifiers defined
// - Identifiers read
// - Identifiers written
// External reads and writes can be calculated from that list. Dependencies
// from sub-blocks are carried upwards/outwards.
//
// Representing identifiers:
// Name [string] this is unique within the block
// Type [ASTType]
package main

import (
	"go/ast"
	"reflect"
)

type DependencyPass struct {
	BasePass
}

type DependencyPassData struct {
	defines []*ast.Ident
	reads   []*ast.Ident
	writes  []*ast.Ident
}

func (d *DependencyPassData) FillIn(isWrite bool, node ast.Node) (err error) {
	return
}

func NewDependencyPassData() *DependencyPassData {
	return &DependencyPassData{}
}

type DependencyPassVisitor struct {
	cur  *BasicBlock
	pass *DependencyPass
	c    *Compiler
}

func (v DependencyPassVisitor) Visit(node ast.Node) (w ast.Visitor) {
	// Get the closest enclosing basic block for this node
	b, ok := v.c.GetPassResult(BasicBlockPassType, node).(*BasicBlock)
	var dataBlock *DependencyPassData
	if !ok {
		b = v.cur
		dataBlock = b.Get(DependencyPassType).(*DependencyPassData)
	} else {
		dataBlock = NewDependencyPassData()
		b.Set(DependencyPassType, dataBlock)
		v.cur = b
	}
	if node == nil {
		// post-order actions (all sub-nodes have been visited)
		//
		// Merge the sub-node read/write list with the current read/write list, and
		// add the block to the current block's child list
		return v
	}

	b.Print("start", reflect.TypeOf(node), node)
	switch t := node.(type) {
	case *ast.CallExpr:
		// tag this node with a reference to the surrounding BasicBlock so we can
		// fill in additional reads/writes once we resolve all functions
		b.Print("+ tagged for later pass")
		v.pass.SetResult(t, b)
	}

	// Analyze all AST nodes for reads, writes and defines
	switch t := node.(type) {
	case *ast.AssignStmt:
		b.Print("- Writes", t.Lhs)
		b.Print("- Reads", t.Rhs)
		return nil // don't go down these branches
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
	return BasicBlockPassMode
}

func (pass *DependencyPass) GetDependencies() []PassType {
	return []PassType{BasicBlockPassType}
}

func MergeDependenciesUpwards(child *BasicBlock, parent *BasicBlock) {

}

func (pass *DependencyPass) RunBasicBlockPass(block *BasicBlock, c *Compiler) (modified bool, err error) {
	v := DependencyPassVisitor{cur: &BasicBlock{}, c: c, pass: pass}
	ast.Walk(v, block.node)

	// Merge this block into the parent
	return
}
