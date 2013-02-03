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
	"go/token"
	"reflect"
)

type DependencyPass struct {
	BasePass
}

type DependencyType uint

const (
	ReadOnly DependencyType = iota
	WriteFirst
	ReadWrite
)

type AccessType uint

const (
	ReadAccess AccessType = iota
	WriteAccess
)

var accessTypeString = map[AccessType]string{
	ReadAccess:  "\x1b[34mReadAccess\x1b[0m",
	WriteAccess: "\x1b[31mWriteAccess\x1b[0m",
}

type DependencyPassData struct {
	defines  map[string]ast.Expr       // name: type expression
	accesses map[string]DependencyType // name: read-only/write-first/read-write
}

func (d *DependencyPassData) FillIn(isWrite bool, node ast.Node) (err error) {
	return
}

func NewDependencyPassData() *DependencyPassData {
	return &DependencyPassData{
		defines:  make(map[string]ast.Expr),
		accesses: make(map[string]DependencyType),
	}
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
	// Helper functions
	Define := func(ident string, expr ast.Expr) {
		dataBlock.defines[ident] = expr
		b.Printf("Defined %s = %T %+v", ident, expr, expr)
	}

	Access := func(ident string, t AccessType) {
		if ident == "_" {
			return
		}
		if prev, ok := dataBlock.accesses[ident]; ok {
			// upgrade the previous access
			if prev == ReadOnly && t == WriteAccess {
				dataBlock.accesses[ident] = ReadWrite
			}
		} else {
			if t == ReadAccess {
				dataBlock.accesses[ident] = ReadOnly
			} else if t == WriteAccess {
				dataBlock.accesses[ident] = WriteFirst
			}
		}
		b.Print("Accessed", ident, accessTypeString[t])
	}

	var AccessExpr func(expr ast.Expr, t AccessType)
	AccessExpr = func(expr ast.Expr, t AccessType) {
		// recursively fill in accesses for an expression
		switch e := expr.(type) {
		case *ast.Ident:
			Access(e.Name, t)
		case *ast.BinaryExpr:
			AccessExpr(e.X, ReadAccess)
			AccessExpr(e.Y, ReadAccess)
		case *ast.CallExpr:
			AccessExpr(e.Fun, ReadAccess)
			for _, funcArg := range e.Args {
				AccessExpr(funcArg, ReadAccess)
			}

		}
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

	// Analyze all AST nodes for definitions and accesses
	switch t := node.(type) {
	case *ast.FuncDecl:
		// defines arguments, named return types
		b.Printf("Func: %+v -> %+v", t.Type.Params, t.Type.Results)
		var defines []*ast.Field
		if t.Type.Params != nil {
			defines = append(defines, t.Type.Params.List...)
		}
		if t.Type.Results != nil {
			defines = append(defines, t.Type.Results.List...)
		}
		for _, paramGroup := range defines {
			for _, param := range paramGroup.Names {
				Define(param.Name, paramGroup.Type)
			}
		}
		return nil
	case *ast.ForStmt:
		b.Printf("For %+v", t.Init)
		// TODO: t.Init defines
		AccessExpr(t.Cond, ReadAccess)
		// TODO: t.Post accesses (writes??)
		//AccessExpr(t.Post, WriteAccess)
		return nil
	case *ast.RangeStmt:
		// only := range
		if t.Tok == token.DEFINE {
			if key, ok := t.Key.(*ast.Ident); ok {
				// TODO: no support for maps, assumes key is always an array index
				Define(key.Name, &ast.Ident{Name: "int"})
			}
			if val, ok := t.Value.(*ast.Ident); ok {
				Define(val.Name, t.X)
			}
		}
		// reads
		AccessExpr(t.X, ReadAccess)
		return nil
	case *ast.IfStmt:
		// descend into t.Init, t.Cond
	case *ast.GenDecl:
		// defines
	case *ast.UnaryExpr:
		AccessExpr(t.X, ReadAccess)
		return nil
	case *ast.BinaryExpr:
		// read only??
		AccessExpr(t.X, ReadAccess)
		AccessExpr(t.Y, ReadAccess)
		return nil
	case *ast.AssignStmt:
		// a[idx], x[idx] = b+c+d, idx
		// writes: a, x reads: idx, b, c, d
		if t.Tok == token.DEFINE {
			for i, expr := range t.Lhs {
				if new, ok := expr.(*ast.Ident); ok {
					Define(new.Name, t.Rhs[i])
				}
			}
		}
		// assignment
		for _, expr := range t.Lhs {
			AccessExpr(expr, WriteAccess)
		}
		for _, expr := range t.Rhs {
			AccessExpr(expr, ReadAccess)
		}

		return nil // don't go down these branches
	case *ast.ExprStmt, *ast.BlockStmt:
		// go down the block
	default:
		b.Printf("\x1b[33mUnknown node\x1b[0m %T %v", t, t)
		return nil // by default don't decend down un-implemented branches
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
	return []PassType{BasicBlockPassType, DefinedTypesPassType}
}

func MergeDependenciesUpwards(child *BasicBlock, parent *BasicBlock) {

}

func (pass *DependencyPass) RunBasicBlockPass(block *BasicBlock, c *Compiler) (modified bool, err error) {
	v := DependencyPassVisitor{cur: &BasicBlock{}, c: c, pass: pass}
	ast.Walk(v, block.node)

	// Merge this block into the parent
	return
}
