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
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
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

type AccessPassData struct {
	//accesses map[string]interface{}    // name.a.b : AccessType or MemoryAccessGroup
	defines  map[string]ast.Expr // name: type expression
	accesses []IdentifierGroup   // read/write of an identifier
}

func (d *AccessPassData) FillIn(isWrite bool, node ast.Node) (err error) {
	return
}

func NewAccessPassData() *AccessPassData {
	return &AccessPassData{
		defines:  make(map[string]ast.Expr),
		accesses: make([]IdentifierGroup, 1),
	}
}

type AccessPassVisitor struct {
	cur       *BasicBlock
	dataBlock *AccessPassData
	pass      *AccessPass
	c         *Compiler
}

func (v AccessPassVisitor) Done(block *BasicBlock) (modified bool, err error) {
	dataBlock := v.dataBlock

	block.Print("== Defines ==")
	for ident, expr := range dataBlock.defines {
		block.Printf("%s = %+v", ident, expr)
	}
	block.Print("== Accesses ==")
	for _, access := range dataBlock.accesses {
		block.Printf("%+v = %s", access.group, accessTypeString[access.t])
	}
	MergeDependenciesUpwards(block)
	return
}

func (v AccessPassVisitor) Visit(node ast.Node) (w BasicBlockVisitor) {
	// Get the closest enclosing basic block for this node
	dataBlock := v.dataBlock
	b := v.cur
	// Helper functions
	Define := func(ident string, expr ast.Expr) {
		if ident == "_" {
			return
		}
		dataBlock.defines[ident] = expr
		b.Printf("Defined %s = %T %+v", ident, expr, expr)
	}

	RecordAccess := func(ident *IdentifierGroup, t AccessType) {
		if ident.group[0].id == "_" {
			return
		}
		dataBlock.accesses = append(dataBlock.accesses, *ident)
		b.Print("Accessed: %+v", ident)
	}

	// TOOD: finish this
	AccessIdentBuild := func(group *IdentifierGroup, expr ast.Expr) {
		var ident Identifier
		switch t := expr.(type) {
		case *ast.Ident:
			ident.id = t.Name
		}
		group.group = append(group.group, ident)
	}

	var AccessExpr func(expr ast.Expr, t AccessType)
	AccessExpr = func(expr ast.Expr, t AccessType) {
		// recursively fill in accesses for an expression
		switch e := expr.(type) {
		case *ast.BinaryExpr:
			AccessExpr(e.X, ReadAccess)
			AccessExpr(e.Y, ReadAccess)
		case *ast.CallExpr:
			AccessExpr(e.Fun, ReadAccess)
			for _, funcArg := range e.Args {
				AccessExpr(funcArg, ReadAccess)
			}
		// These three expression types form the basis of a memory access.
		// ast.Ident
		// ast.IndexExpr
		// ast.SelectorExpr
		case *ast.IndexExpr:
			// a[idx+1]
			// e.X = a
			// e.Index = idx+1
			// TODO: more granular - treat each index as unique
			idGroup := &IdentifierGroup{}
			AccessIdentBuild(idGroup, e.X)
			RecordAccess(idGroup, t)
			//AccessExpr(e.X, t)
			idGroup = &IdentifierGroup{}
			AccessIdentBuild(idGroup, e.Index)
			RecordAccess(idGroup, ReadAccess)
			//AccessExpr(e.Index, ReadAccess)
		case *ast.SelectorExpr:
			// x.y.z expressions
			// e.X = x.y
			// e.Sel = y
			// TODO: more granular level of read controls
			// TODO: what about x.y[a]? wouldn't pick up read to [a]
			// Support:
			// a[idx].b.c.[idx2].e
			switch e.X.(type) {
			case *ast.Ident:
				// Build a joint ast.Ident "a.b"
				//e.Sel expr.Name
			}
			AccessExpr(e.X, t)
		case *ast.Ident:
			AccessExpr(e, t)
		case *ast.BasicLit:
			// ignore, builtin constant
		default:
			fmt.Printf("Unknown access expression %T\n", expr)
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
		// fallthrough to descend into t.Init, t.Cond
	case *ast.GenDecl:
		// defines
	case *ast.ValueSpec:
		// defines
		for _, name := range t.Names {
			Define(name.Name, t.Type)
		}
		// reads
		for _, expr := range t.Values {
			AccessExpr(expr, ReadAccess)
		}
		return nil
	case *ast.UnaryExpr:
		AccessExpr(t.X, ReadAccess)
		return nil
	case *ast.IncDecStmt:
		AccessExpr(t.X, ReadAccess)
		AccessExpr(t.X, WriteAccess)
		return nil
	case *ast.BinaryExpr:
		// read only??
		AccessExpr(t.X, ReadAccess)
		AccessExpr(t.Y, ReadAccess)
		return nil
	case *ast.AssignStmt:
		// a[idx], x[idx] = b+c+d, idx
		// writes: a, x reads: idx, b, c, d
		switch t.Tok {
		case token.DEFINE:
			for i, expr := range t.Lhs {
				if new, ok := expr.(*ast.Ident); ok {
					Define(new.Name, t.Rhs[i])
				}
			}
		case token.ASSIGN:
		default:
			// x += 1, etc
			for _, expr := range t.Lhs {
				AccessExpr(expr, ReadAccess)
			}
		}

		// assignment (read LHS index, read RHS, write LHS)
		for _, expr := range t.Rhs {
			AccessExpr(expr, ReadAccess)
		}
		for _, expr := range t.Lhs {
			b.Printf("write %T %+v", expr, expr)
			AccessExpr(expr, WriteAccess)
		}
		return nil // don't go down these branches
	case *ast.SendStmt:
		AccessExpr(t.Value, ReadAccess)
		AccessExpr(t.Chan, WriteAccess)
		return nil
	case *ast.CallExpr:
		AccessExpr(t.Fun, ReadAccess)
		for _, expr := range t.Args {
			AccessExpr(expr, ReadAccess)
		}
		return nil
	case *ast.BranchStmt:
		// ignore break/continue/goto/fallthrough
		return nil
	case *ast.ExprStmt, *ast.BlockStmt, *ast.ReturnStmt, *ast.DeclStmt,
		*ast.GoStmt:
		// go down the block

	default:
		b.Printf("\x1b[33mUnknown node\x1b[0m %T %+v", t, t)
		return nil // by default don't decend down un-implemented branches
	}
	return v
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

func MergeDependenciesUpwards(child *BasicBlock) {
	// TODO: merge reads/writes of identifiers outside this scope
}

func (pass *AccessPass) RunBasicBlockPass(block *BasicBlock, c *Compiler) BasicBlockVisitor {
	dataBlock := NewAccessPassData()
	block.Set(AccessPassType, dataBlock)
	return AccessPassVisitor{cur: block, dataBlock: dataBlock, c: c, pass: pass}
}
