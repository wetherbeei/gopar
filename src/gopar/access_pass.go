// Analyze the data dependencies between "basic blocks" (for loops + functions)
// Any dependencies within blocks don't matter, only those that cross the
// boundaries of loops and function calls. Each basic block records:
// - New identifiers defined
// - Identifiers read
// - Identifiers written
// External reads and writes can be calculated from that list. Dependencies
// from sub-blocks are carried upwards/outwards.

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
)

type AccessPass struct {
	BasePass
}

type AccessType uint

const (
	ReadAccess AccessType = iota
	WriteAccess
)

var accessTypeString = map[AccessType]string{
	ReadAccess:  "\x1b[34mReadAccess\x1b[0m",
	WriteAccess: "\x1b[31mWriteAccess\x1b[0m",
}

type Identifier struct {
	id        string
	isIndexed bool   // true if this is an indexed identifier
	index     string // single identifier [idx] support for now
}

// Represents a.b.c[expr].d accesses
type IdentifierGroup struct {
	t     AccessType
	group []Identifier // [a, b, c[affine expr], d]
}

func (ig *IdentifierGroup) String() string {
	var buffer bytes.Buffer
	for _, i := range ig.group {
		buffer.WriteString(i.id)
		if i.isIndexed {
			buffer.WriteString("[")
			buffer.WriteString(i.index)
			buffer.WriteString("]")
		}
		buffer.WriteString(".")
	}
	buffer.WriteString(accessTypeString[ig.t])
	return buffer.String()
}

type AccessPassData struct {
	//accesses map[string]interface{}    // name.a.b : AccessType or MemoryAccessGroup
	defines  map[string]ast.Expr // name: type expression
	accesses []IdentifierGroup   // read/write of an identifier
}

func NewAccessPassData() *AccessPassData {
	return &AccessPassData{
		defines:  make(map[string]ast.Expr),
		accesses: make([]IdentifierGroup, 0),
	}
}

type AccessPassVisitor struct {
	cur       *BasicBlock
	dataBlock *AccessPassData
	pass      *AccessPass
	p         *Package
}

func (v AccessPassVisitor) Done(block *BasicBlock) (modified bool, err error) {
	dataBlock := v.dataBlock

	block.Print("== Defines ==")
	for ident, expr := range dataBlock.defines {
		block.Printf("%s = %T %+v", ident, expr, expr)
	}
	block.Print("== Accesses ==")
	for _, access := range dataBlock.accesses {
		block.Printf(access.String())
	}

	return
}

func (pass *AccessPass) ParseBasicBlock(node ast.Node, p *Package) {
	// Get the closest enclosing basic block for this node
	b := pass.GetCompiler().GetPassResult(BasicBlockPassType, node).(*BasicBlock)
	dataBlock := NewAccessPassData()
	b.Set(AccessPassType, dataBlock)

	b.Printf("start %T %+v", node, node)

	// Helper functions.
	// Define adds the identifier as being defined in this block
	Define := func(ident string, expr ast.Expr) {
		if ident == "_" {
			return
		}
		dataBlock.defines[ident] = expr
		b.Printf("Defined %s = %T %+v", ident, expr, expr)
	}

	// Access recording:
	// - RecordAccess records the final identifier group
	// - AccessIdentBuild takes an identifier group and an expression and
	//     recursively builds the group out of the expression
	// - AccessExpr 
	RecordAccess := func(ident *IdentifierGroup, t AccessType) {
		ident.t = t
		if ident.group[0].id == "_" {
			return
		}
		dataBlock.accesses = append(dataBlock.accesses, *ident)
		b.Printf("Accessed: %+v", ident)
	}

	// Support:
	// a[idx].b
	// a.b[idx].c
	var AccessIdentBuild func(group *IdentifierGroup, expr ast.Expr)
	var AccessExpr func(expr ast.Expr, t AccessType)

	AccessIdentBuild = func(group *IdentifierGroup, expr ast.Expr) {
		var ident Identifier
		switch t := expr.(type) {
		case *ast.Ident:
			ident.id = t.Name
		case *ast.SelectorExpr:
			// x.y.z expressions
			// e.X = x.y
			// e.Sel = z
			ident.id = t.Sel.Name
			AccessIdentBuild(group, t.X)
		case *ast.IndexExpr:
			// a[idx][x]
			// a[idx+1]
			// e.X = a
			// e.Index = idx+1
			//AccessIdent(t.Index, ReadAccess)
			switch x := t.X.(type) {
			case *ast.Ident:
				ident.id = x.Name
			default:
				b.Print("Unresolved array expression %T %+v", x, x)
			}
			AccessExpr(t.Index, ReadAccess)
			switch i := t.Index.(type) {
			case *ast.Ident:
				ident.index = i.Name
				ident.isIndexed = true
			default:
				// can't resolve array access, record for the entire array
				b.Printf("Unresolved array access %T [%+v]", i, i)
			}
		default:
			b.Printf("Unknown ident %T %+v", t, t)
		}
		group.group = append(group.group, ident)
	}

	AccessIdent := func(expr ast.Expr, t AccessType) {
		ig := &IdentifierGroup{}
		AccessIdentBuild(ig, expr)
		RecordAccess(ig, t)
	}

	// This is called recursively, starting with the block level
	// Don't recurse down into other BasicBlock statements, instead manually
	// call pass.RunBasicBlockPass()
	AccessExpr = func(node ast.Node, t AccessType) {
		if isBasicBlockNode(node) {
			pass.ParseBasicBlock(node, p)
			return // don't desend into the block
		}
		// recursively fill in accesses for an expression
		switch e := node.(type) {
		case *ast.BinaryExpr:
			b.Printf("BinaryExpr %T %+v , %T %+v", e.X, e.X, e.Y, e.Y)
			AccessExpr(e.X, ReadAccess)
			AccessExpr(e.Y, ReadAccess)
		case *ast.CallExpr:
			for _, funcArg := range e.Args {
				AccessExpr(funcArg, ReadAccess)
			}
		// These three expression types form the basis of a memory access.
		case *ast.SelectorExpr, *ast.IndexExpr, *ast.Ident:
			AccessIdent(e, t)
		case *ast.UnaryExpr:
			// <-x
			AccessExpr(e.X, ReadAccess)
		case *ast.BasicLit:
			// ignore, builtin constant
		case *ast.ArrayType, *ast.ChanType:
			// ignore, type expressions
		default:

			fmt.Printf("Unknown access expression %T %+v\n", expr, expr)

		}
	}
	return
}

func NewAccessPass() *AccessPass {
	return &AccessPass{
		BasePass: NewBasePass(),
	}
}

func (pass *AccessPass) GetPassType() PassType {
	return AccessPassType
}

func (pass *AccessPass) GetPassMode() PassMode {
	return BasicBlockPassMode
}

func (pass *AccessPass) GetDependencies() []PassType {
	return []PassType{BasicBlockPassType, DefinedTypesPassType}
}

func (pass *AccessPass) RunFunctionPass(fun *ast.FuncDecl, p *Package) (modified bool, err error) {
	pass.ParseBasicBlock(fun, p)
	return
}
