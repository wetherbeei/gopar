// Basic Block creator
//
// Create a tree of basic blocks for variable scoping

package main

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"
)

// TODO: These functions have a corresponding GPU implementation
var builtinTranslated map[string]bool = map[string]bool{
// Detect constant make lengths and privatize an array
// "make": bool
}

// A basic block represents a Go scope (function, for, range, if, switch).
type BasicBlock struct {
	depth    int
	node     ast.Node
	parent   *BasicBlock
	children []*BasicBlock
	data     map[PassType]interface{}
}

func NewBasicBlock(node ast.Node, parent *BasicBlock) *BasicBlock {
	b := &BasicBlock{
		node:   node,
		parent: parent,
		data:   make(map[PassType]interface{}),
	}
	if parent != nil {
		b.depth = parent.depth + 1
		parent.children = append(parent.children, b)
	}
	return b
}

func (b *BasicBlock) Get(t PassType) interface{} {
	return b.data[t]
}

func (b *BasicBlock) Set(t PassType, i interface{}) {
	b.data[t] = i
}

// Print a message at this block level
func (b BasicBlock) Print(args ...interface{}) {
	args = append(args, 0)
	copy(args[1:], args[0:])
	args[0] = 1
	args[0] = strings.Repeat("  ", b.depth)
	fmt.Println(args...)
}

type BasicBlockPass struct {
	BasePass
}

type BasicBlockPassVisitor struct {
	cur  *BasicBlock
	pass *BasicBlockPass
	c    *Compiler
}

func (v BasicBlockPassVisitor) Visit(node ast.Node) (w ast.Visitor) {
	block := v.cur
	if node == nil {
		// post-order actions (all sub-nodes have been visited)
		block.Print("done", reflect.TypeOf(node))
		return v
	}
	block.Print("start", reflect.TypeOf(node), node)
	switch t := node.(type) {
	case *ast.FuncLit, *ast.ForStmt, *ast.RangeStmt, *ast.FuncDecl:
		block.Print("+ new block", t)
		newBlock := NewBasicBlock(node, block)
		v.cur = newBlock
		v.pass.SetResult(t, newBlock)
	}
	return v
}

func NewBasicBlockPass() *BasicBlockPass {
	return &BasicBlockPass{
		BasePass: NewBasePass(),
	}
}

func (pass *BasicBlockPass) GetPassType() PassType {
	return BasicBlockPassType
}

func (pass *BasicBlockPass) GetPassMode() PassMode {
	return FunctionPassMode
}

func (pass *BasicBlockPass) GetDependencies() []PassType {
	return []PassType{}
}

func (pass *BasicBlockPass) RunFunctionPass(function *ast.FuncDecl, c *Compiler) (modified bool, err error) {
	v := BasicBlockPassVisitor{cur: NewBasicBlock(function, nil), c: c, pass: pass}
	ast.Walk(v, function)
	return
}
