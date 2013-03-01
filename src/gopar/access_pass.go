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
	"math/rand"
	"strconv"
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

type AccessExprFn func(node ast.Node, t AccessType)

func AccessIdentBuild(group *IdentifierGroup, expr ast.Node, fn AccessExprFn) {
	var ident Identifier
	switch t := expr.(type) {
	case *ast.Ident:
		ident.id = t.Name
	case *ast.SelectorExpr:
		// x.y.z expressions
		// e.X = x.y
		// e.Sel = z
		ident.id = t.Sel.Name
		AccessIdentBuild(group, t.X, fn)
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
			fmt.Printf("Unresolved array expression %T %+v", x, x)
		}
		if fn != nil {
			fn(t.Index, ReadAccess)
		}
		switch i := t.Index.(type) {
		case *ast.Ident:
			ident.index = i.Name
			ident.isIndexed = true
		default:
			// can't resolve array access, record for the entire array
			fmt.Printf("Unresolved array access %T [%+v]\n", i, i)
		}
	default:
		fmt.Printf("Unknown expression %T %+v\n", t, t)
	}
	group.group = append(group.group, ident)
}

func (pass *AccessPass) ParseBasicBlock(blockNode ast.Node, p *Package) {
	// Get the closest enclosing basic block for this node
	b := pass.GetCompiler().GetPassResult(BasicBlockPassType, blockNode).(*BasicBlock)
	dataBlock := NewAccessPassData()
	b.Set(AccessPassType, dataBlock)
	b.Printf("\x1b[32;1mBasicBlock\x1b[0m %T %+v", blockNode, blockNode)
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

	var AccessExpr AccessExprFn

	AccessIdent := func(expr ast.Node, t AccessType) {
		ig := &IdentifierGroup{}
		AccessIdentBuild(ig, expr, AccessExpr)
		RecordAccess(ig, t)
	}

	// This is called recursively, starting with the block level
	// Don't recurse down into other BasicBlock statements, instead manually
	// call pass.RunBasicBlockPass()
	AccessExpr = func(node ast.Node, t AccessType) {
		if node == nil {
			return
		}
		if isBasicBlockNode(node) && node != blockNode {
			pass.ParseBasicBlock(node, p)
			return // don't desend into the block
		}
		b.Printf("start %T %+v", node, node)
		// recursively fill in accesses for an expression
		switch e := node.(type) {
		// These three expression types form the basis of a memory access.
		case *ast.SelectorExpr, *ast.IndexExpr, *ast.Ident:
			AccessIdent(e, t)

		// Statement/expression blocks
		case *ast.BinaryExpr:
			b.Printf("BinaryExpr %T %+v , %T %+v", e.X, e.X, e.Y, e.Y)
			AccessExpr(e.X, ReadAccess)
			AccessExpr(e.Y, ReadAccess)
		case *ast.UnaryExpr:
			// <-x
			AccessExpr(e.X, ReadAccess)
		case *ast.BasicLit:
			// ignore, builtin constant
		case *ast.ArrayType, *ast.ChanType:
			// ignore, type expressions
		case *ast.FuncDecl:
			AccessExpr(e.Type, ReadAccess)
			AccessExpr(e.Body, ReadAccess)
		case *ast.FuncLit: // we don't support closures, but gather access info
			AccessExpr(e.Type, ReadAccess)
			AccessExpr(e.Body, ReadAccess)
		case *ast.FuncType:
			// defines arguments, named return types
			b.Printf("Func: %+v -> %+v", e.Params, e.Results)
			var defines []*ast.Field
			if e.Params != nil {
				defines = append(defines, e.Params.List...)
			}
			if e.Results != nil {
				defines = append(defines, e.Results.List...)
			}
			for _, paramGroup := range defines {
				for _, param := range paramGroup.Names {
					Define(param.Name, paramGroup.Type)
				}
			}
		case *ast.ForStmt:
			b.Printf("For %+v; %+v; %+v", e.Init, e.Cond, e.Post)
			AccessExpr(e.Init, ReadAccess)
			AccessExpr(e.Cond, ReadAccess)

			AccessExpr(e.Body, ReadAccess)

			AccessExpr(e.Post, ReadAccess)
		case *ast.RangeStmt:
			// only := range
			if e.Tok == token.DEFINE {
				if key, ok := e.Key.(*ast.Ident); ok {
					// TODO: no support for maps, assumes key is always an array index
					Define(key.Name, &ast.Ident{Name: "int"})
				}
				if val, ok := e.Value.(*ast.Ident); ok {
					Define(val.Name, e.X)
				}
			}
			// reads
			AccessExpr(e.X, ReadAccess)
			AccessExpr(e.Body, ReadAccess)
		case *ast.IfStmt:
			// fallthrough to descend into e.Init, e.Cond
			AccessExpr(e.Init, ReadAccess)
			AccessExpr(e.Cond, ReadAccess)
			AccessExpr(e.Body, ReadAccess)

			AccessExpr(e.Else, ReadAccess)
		case *ast.DeclStmt:
			AccessExpr(e.Decl, ReadAccess)
		case *ast.GenDecl:
			for _, s := range e.Specs {
				AccessExpr(s, ReadAccess)
			}
		case *ast.ValueSpec:
			// defines
			for _, name := range e.Names {
				Define(name.Name, e.Type)
			}
			// reads
			for _, expr := range e.Values {
				AccessExpr(expr, ReadAccess)
			}
		case *ast.IncDecStmt:
			AccessExpr(e.X, ReadAccess)
			AccessExpr(e.X, WriteAccess)
		case *ast.AssignStmt:
			// a[idx], x[idx] = b+c+d, idx
			// writes: a, x reads: idx, b, c, d
			switch e.Tok {
			case token.DEFINE:
				for i, expr := range e.Lhs {
					if new, ok := expr.(*ast.Ident); ok {
						Define(new.Name, e.Rhs[i])
					}
				}
			case token.ASSIGN:
			default:
				// x += 1, etc
				for _, expr := range e.Lhs {
					AccessExpr(expr, ReadAccess)
				}
			}

			// assignment (read LHS index, read RHS, write LHS)
			for _, expr := range e.Rhs {
				AccessExpr(expr, ReadAccess)
			}
			for _, expr := range e.Lhs {
				b.Printf("write %T %+v", expr, expr)
				AccessExpr(expr, WriteAccess)
			}
		case *ast.SendStmt:
			AccessExpr(e.Value, ReadAccess)
			AccessExpr(e.Chan, WriteAccess)
		case *ast.CallExpr:
			for _, expr := range e.Args {
				AccessExpr(expr, ReadAccess)
			}
			switch f := e.Fun.(type) {
			case *ast.Ident:
				// Fill in additional reads/writes once we resolve all functions.
				// Insert a placeholder access that will be removed later
				if p.Lookup(f.Name) != nil {
					placeholder := strconv.FormatInt(rand.Int63(), 10)
					placeholderIdent := &ast.Ident{Name: placeholder}
					AccessExpr(placeholderIdent, ReadAccess)
					pass.SetResult(e, placeholderIdent)
					b.Printf("\x1b[33m>> %s\x1b[0m use in later pass: %s", placeholder, f.Name)
				} else {
					// we can't see inside the function, assume all of the args are
					// written to
					for _, expr := range e.Args {
						AccessExpr(expr, WriteAccess)
					}
				}
				return
			case *ast.FuncLit:
				// gather external accesses, and search inside closure, but don't bother
				// propogating the aliased arguments
				AccessExpr(e.Fun, ReadAccess)
			}
			b.Printf("\x1b[33mUnsupported CallExpr %T\x1b[0m", e.Fun)
		case *ast.BranchStmt:
			// ignore break/continue/goto/fallthrough
		case *ast.ExprStmt:
			AccessExpr(e.X, ReadAccess)
		case *ast.BlockStmt:
			for _, l := range e.List {
				AccessExpr(l, ReadAccess)
			}
		case *ast.ReturnStmt:
			for _, l := range e.Results {
				AccessExpr(l, ReadAccess)
			}
		case *ast.GoStmt:
			AccessExpr(e.Call, ReadAccess)
		default:
			b.Printf("\x1b[33mUnknown node\x1b[0m %T %+v", e, e)
		}
	}
	AccessExpr(blockNode, ReadAccess)
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
	return FunctionPassMode
}

func (pass *AccessPass) GetDependencies() []PassType {
	return []PassType{BasicBlockPassType, DefinedTypesPassType}
}

func (pass *AccessPass) RunFunctionPass(fun *ast.FuncDecl, p *Package) (modified bool, err error) {
	pass.ParseBasicBlock(fun, p)
	return
}
