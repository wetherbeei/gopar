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

var dependencyTypeString = map[DependencyType]string{
	ReadOnly:   "\x1b[34mReadOnly\x1b[0m",
	WriteFirst: "\x1b[31mWriteFirst\x1b[0m",
	ReadWrite:  "\x1b[32mReadWrite\x1b[0m",
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
	cur       *BasicBlock
	dataBlock *DependencyPassData
	pass      *DependencyPass
	c         *Compiler
}

func (v DependencyPassVisitor) Visit(node ast.Node) (w ast.Visitor) {
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
		// break/continue/goto/fallthrough
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

func MergeDependenciesUpwards(child *BasicBlock) {
	// TODO: merge reads/writes of identifiers outside this scope
}

func (pass *DependencyPass) RunBasicBlockPass(block *BasicBlock, c *Compiler) (modified bool, err error) {
	dataBlock := NewDependencyPassData()
	block.Set(DependencyPassType, dataBlock)
	v := DependencyPassVisitor{cur: block, dataBlock: dataBlock, c: c, pass: pass}
	ast.Walk(v, block.node)

	block.Print("== Defines ==")
	for ident, expr := range dataBlock.defines {
		block.Printf("%s = %+v", ident, expr)
	}
	block.Print("== Accesses ==")
	for ident, t := range dataBlock.accesses {
		block.Printf("%s = %s", ident, dependencyTypeString[t])
	}
	MergeDependenciesUpwards(block)
	return
}
