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
	c         *Compiler
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

func (v AccessPassVisitor) Visit(node ast.Node) (w BasicBlockVisitor) {
	// Get the closest enclosing basic block for this node
	dataBlock := v.dataBlock
	b := v.cur
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

	AccessExpr = func(expr ast.Expr, t AccessType) {
		// recursively fill in accesses for an expression
		switch e := expr.(type) {
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

	if node == nil {
		// post-order actions (all sub-nodes have been visited)
		return v
	}

	b.Printf("start %T %+v", node, node)

	// Analyze all AST nodes for definitions and accesses
	switch t := node.(type) {
	case *ast.FuncDecl, *ast.FuncLit:
		// pass - go into FuncType and the body BlockStmt
	case *ast.FuncType:
		// defines arguments, named return types
		b.Printf("Func: %+v -> %+v", t.Params, t.Results)
		var defines []*ast.Field
		if t.Params != nil {
			defines = append(defines, t.Params.List...)
		}
		if t.Results != nil {
			defines = append(defines, t.Results.List...)
		}
		for _, paramGroup := range defines {
			for _, param := range paramGroup.Names {
				Define(param.Name, paramGroup.Type)
			}
		}
		return nil
	case *ast.ForStmt:
		b.Printf("For %+v; %+v; %+v", t.Init, t.Cond, t.Post)
		// TODO: t.Init defines
		//AccessExpr(t.Cond, ReadAccess)
		// TODO: t.Post accesses (writes??)
		//AccessExpr(t.Post, WriteAccess)
		//return nil
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
		for _, expr := range t.Args {
			AccessExpr(expr, ReadAccess)
		}
		switch f := t.Fun.(type) {
		case *ast.FuncLit:
			// go down a FuncLit branch (anonymous closure)
		case *ast.Ident:
			// fill in additional reads/writes once we resolve all functions
			b.Printf("\x1b[33m+\x1b[0m use in later pass: %s", f.Name)
			return nil
		default:
			b.Printf("\x1b[33mUnknown CallExpr %T\x1b[0m", t.Fun)
		}
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

func (pass *AccessPass) RunBasicBlockPass(block *BasicBlock, p *Package) BasicBlockVisitor {
	dataBlock := NewAccessPassData()
	block.Set(AccessPassType, dataBlock)
	return AccessPassVisitor{cur: block, dataBlock: dataBlock, pass: pass}
}
