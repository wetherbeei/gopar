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

type ReferenceType uint

const (
	NoReference ReferenceType = iota
	Dereference               // *a
	AddressOf                 // &a
)

type Identifier struct {
	id        string
	refType   ReferenceType
	isIndexed bool   // true if this is an indexed identifier
	index     string // single identifier [idx] support for now
}

func (i *Identifier) Equals(i2 Identifier) bool {
	if i.id != i2.id {
		return false // selector doesn't match
	}
	if i.isIndexed != i2.isIndexed {
		return false // one is indexed, the other isn't
	}
	if i.isIndexed && i.index != i2.index {
		return false // index doesn't match
	}
	if i.refType != i2.refType {
		return false // reference types don't match
	}
	return true
}

// Represents a.b.c[expr].d accesses
type IdentifierGroup struct {
	t     AccessType
	group []Identifier // [a, b, c[affine expr], d]
}

func (ig *IdentifierGroup) String() string {
	var buffer bytes.Buffer
	for _, i := range ig.group {
		if i.refType == Dereference {
			buffer.WriteString("*")
		} else if i.refType == AddressOf {
			buffer.WriteString("&")
		}
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
	defines  map[string]Type   // name: type
	accesses []IdentifierGroup // read/write of an identifier
}

func NewAccessPassData() *AccessPassData {
	return &AccessPassData{
		defines:  make(map[string]Type),
		accesses: make([]IdentifierGroup, 0),
	}
}

type AccessExprFn func(node ast.Node, t AccessType)

func AccessIdentBuild(group *IdentifierGroup, expr ast.Node, fn AccessExprFn) {
	var ident Identifier
	// deal with pointers and addresses first
	switch t := expr.(type) {
	case *ast.StarExpr:
		ident.refType = Dereference
		expr = t.X
	case *ast.UnaryExpr:
		if t.Op == token.AND {
			ident.refType = AddressOf
		}
		expr = t.X
	}
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
		case *ast.SelectorExpr:
			ident.id = x.Sel.Name
			AccessIdentBuild(group, x.X, fn)
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
	case *ast.CompositeLit:
		// ignore, we're building a type expression, no accesses here
	case *ast.CallExpr:
		if fn != nil {
			fn(t, ReadAccess)
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
	pos := p.Location(blockNode.Pos())
	b.Printf("\x1b[32;1mBasicBlock %s:%d\x1b[0m %T %+v", pos.Filename, pos.Line, blockNode, blockNode)
	// Helper functions.
	Resolver := MakeResolver(b, p, pass.compiler)
	// Define adds the identifier as being defined in this block
	Define := func(ident string, expr interface{}) {
		if ident == "_" {
			return
		}
		var typ Type
		switch t := expr.(type) {
		case ast.Expr:
			typ = TypeOfDecl(t, Resolver)
		case Type:
			typ = t
		}
		dataBlock.defines[ident] = typ
		if *verbose {
			b.Printf("Defined %s = %v", ident, typ)
		}
	}

	// Access recording:
	// - RecordAccess records the final identifier group
	// - AccessIdentBuild takes an identifier group and an expression and
	//     recursively builds the group out of the expression
	// - AccessExpr 
	RecordAccess := func(ident *IdentifierGroup, t AccessType) {
		ident.t = t
		switch ident.group[0].id {
		case "_", "true", "false", "iota", "":
			return
		}
		dataBlock.accesses = append(dataBlock.accesses, *ident)
		if *verbose {
			b.Printf("Accessed: %+v", ident)
		}
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
	var previousDecl ast.Expr // for constant defines
	AccessExpr = func(node ast.Node, t AccessType) {
		if node == nil {
			return
		}
		if isBasicBlockNode(node) && node != blockNode {
			pass.ParseBasicBlock(node, p)
			return // don't desend into the block
		}
		if *verbose {
			b.Printf("start %T %+v", node, node)
		}
		// recursively fill in accesses for an expression
		switch e := node.(type) {
		// These three expression types form the basis of a memory access.
		case *ast.SelectorExpr, *ast.IndexExpr, *ast.Ident, *ast.StarExpr:
			AccessIdent(e, t)
		case *ast.UnaryExpr:
			// only decend down a unary if it's for a memory ref &
			if e.Op == token.AND {
				AccessIdent(e, t)
			} else {
				AccessExpr(e.X, t)
			}
		// Statement/expression blocks
		case *ast.ParenExpr:
			AccessExpr(e.X, ReadAccess)
		case *ast.BinaryExpr:
			AccessExpr(e.X, ReadAccess)
			AccessExpr(e.Y, ReadAccess)
		case *ast.BasicLit:
			// ignore, builtin constant
		case *ast.ArrayType, *ast.ChanType, *ast.MapType:
			// ignore, type expressions
		case *ast.CompositeLit:
			for _, i := range e.Elts {
				AccessExpr(i, t)
			}
		case *ast.FuncDecl:
			if e.Recv != nil {
				r := e.Recv.List[0]
				if len(r.Names) > 0 {
					// it's valid to declare a recevier without naming the parameter
					Define(r.Names[0].Name, r.Type)
				}
			}
			AccessExpr(e.Type, ReadAccess)
			if e.Body != nil {
				AccessExpr(e.Body, ReadAccess)
			}
		case *ast.FuncLit: // gather accesses like the call happens here
			AccessExpr(e.Type, ReadAccess)
			AccessExpr(e.Body, ReadAccess)
		case *ast.FuncType:
			// defines arguments, named return types
			var defines []*ast.Field
			if e.Params != nil {
				defines = append(defines, e.Params.List...)
			}
			if e.Results != nil {
				defines = append(defines, e.Results.List...)
			}
			for _, paramGroup := range defines {
				for _, param := range paramGroup.Names {
					// check if the last argument is 
					Define(param.Name, paramGroup.Type)
				}
			}
		case *ast.ForStmt:
			AccessExpr(e.Init, ReadAccess)
			AccessExpr(e.Cond, ReadAccess)

			AccessExpr(e.Body, ReadAccess)

			AccessExpr(e.Post, ReadAccess)
		case *ast.RangeStmt:
			// only := range
			if e.Tok == token.DEFINE {
				rangeTyp := TypeOf(e.X, Resolver)
				if key, ok := e.Key.(*ast.Ident); ok {
					Define(key.Name, rangeTyp.IndexKey())
				}
				if val, ok := e.Value.(*ast.Ident); ok {
					Define(val.Name, rangeTyp.IndexValue())
				}
			}
			// reads from range val
			// TODO: re-enable this access?
			//AccessExpr(e.X, ReadAccess)
			// writes to keys
			AccessExpr(e.Key, WriteAccess)
			if e.Value != nil {
				AccessExpr(e.Value, WriteAccess)
			}
			AccessExpr(e.Body, ReadAccess)
		case *ast.IfStmt:
			// fallthrough to descend into e.Init, e.Cond
			AccessExpr(e.Init, ReadAccess)
			AccessExpr(e.Cond, ReadAccess)
			AccessExpr(e.Body, ReadAccess)

			AccessExpr(e.Else, ReadAccess)
		case *ast.SelectStmt:
			AccessExpr(e.Body, ReadAccess)
		case *ast.SwitchStmt:
			AccessExpr(e.Init, ReadAccess)
			AccessExpr(e.Body, ReadAccess)
		case *ast.TypeSwitchStmt:
			// v := x.(type)
			if e.Init != nil {
				AccessExpr(e.Init, ReadAccess)
			}
			// by default, define the switched identifier "v" to have type of "x"
			AccessExpr(e.Assign, ReadAccess)
			AccessExpr(e.Body, ReadAccess)
		case *ast.CommClause:
			AccessExpr(e.Comm, ReadAccess)
			for _, b := range e.Body {
				AccessExpr(b, ReadAccess)
			}
		case *ast.CaseClause:
			// check if this is part of a TypeSwitchStmt, then define the identifier
			// again (v := x.(type))
			if parentSwitch, ok := b.parent.parent.node.(*ast.TypeSwitchStmt); ok {
				switch assign := parentSwitch.Assign.(type) {
				case *ast.AssignStmt:
					// does this case statement have exactly one type?
					if len(e.List) == 1 {
						// redefine switch variable
						Define(assign.Lhs[0].(*ast.Ident).Name, TypeOfDecl(e.List[0], Resolver))
					}
				}
			}
			for _, c := range e.List {
				AccessExpr(c, ReadAccess)
			}
			for _, b := range e.Body {
				AccessExpr(b, ReadAccess)
			}
		case *ast.DeclStmt:
			AccessExpr(e.Decl, ReadAccess)
		case *ast.GenDecl:
			previousDecl = nil // for constants
			for _, s := range e.Specs {
				AccessExpr(s, ReadAccess)
			}
		case *ast.ValueSpec:
			// defines - these are either typed or constant
			if e.Type != nil {
				for _, name := range e.Names {
					Define(name.Name, e.Type)
				}
			} else {
				// all constants
				for i, name := range e.Names {
					if i < len(e.Values) {
						previousDecl = e.Values[i]
					}
					Define(name.Name, previousDecl)
				}
			}

			// reads
			for _, expr := range e.Values {
				AccessExpr(expr, ReadAccess)
			}
		case *ast.IncDecStmt:
			AccessExpr(e.X, ReadAccess)
			AccessExpr(e.X, WriteAccess)
		case *ast.AssignStmt:
			b.Print("Assigning")
			// a[idx], x[idx] = b+c+d, idx
			// writes: a, x reads: idx, b, c, d
			switch e.Tok {
			case token.DEFINE:
				// multi-assign functions or type conversions
				if len(e.Lhs) != len(e.Rhs) {
					if len(e.Rhs) != 1 {
						b.Printf("ERROR: invalid multi-assign: %d to %d", len(e.Lhs), len(e.Rhs))
					}
					result := TypeOf(e.Rhs[0], Resolver).(*MultiType).Expand()
					b.Print(result)
					for i, lhs := range e.Lhs {
						Define(lhs.(*ast.Ident).Name, result[i])
					}
				} else {
					b.Print("single assign", e.Lhs)
					// assign each individually
					for i, expr := range e.Lhs {
						if new, ok := expr.(*ast.Ident); ok {
							Define(new.Name, TypeOf(e.Rhs[i], Resolver))
						}
					}
				}
			case token.ASSIGN:
				// no defines here, assigns are done below
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
				if *verbose {
					b.Printf("write %T %+v", expr, expr)
				}
				AccessExpr(expr, WriteAccess)
			}
		case *ast.SendStmt:
			AccessExpr(e.Value, ReadAccess)
			AccessExpr(e.Chan, WriteAccess)
		case *ast.CallExpr:
			for _, expr := range e.Args {
				AccessExpr(expr, ReadAccess)
			}
			if *verbose {
				fmt.Printf("%T %+v\n", e.Fun, e.Fun)
			}
			fnTyp, ok := TypeOf(e.Fun, Resolver).(*FuncType)
			// Fill in additional reads/writes once we resolve all functions.
			// Insert a placeholder access that will be removed later
			if ok && fnTyp != nil {
				if fnTyp.body != nil {
					placeholder := strconv.FormatInt(rand.Int63(), 10)
					placeholderIdent := &ast.Ident{Name: placeholder}
					AccessExpr(placeholderIdent, ReadAccess)
					pass.SetResult(e, placeholderIdent)
					if *verbose {
						b.Printf("\x1b[33m>> %s\x1b[0m use in later pass: %s", placeholder, fnTyp.String())
					}
					// if this is a function literal, keep decending down to gather
					// the accesses made
					if lit, ok := e.Fun.(*ast.FuncLit); ok {
						b.Print("Decending down a FuncLit")
						AccessExpr(lit, ReadAccess)
					}
				} else {
					// we can't see inside the function, assume all of the pointer args are
					// written to
					b.Printf("\x1b[33mOpaque function:\x1b[0m %s", fnTyp.String())
					classify := func(arg ast.Node, argTyp Type) {
						if argTyp.PassByValue() {
							AccessExpr(arg, ReadAccess)
						} else {
							AccessExpr(arg, WriteAccess)
						}
					}

					// TODO/safety bug: what if an argument is a function?
					// If we know the exact function foo(math.Mul), then propogate th
					// accesses. If the function is just a pointer and we can't see the
					// function definition, assume it modifies some external state.
					//
					// Is this handled by setting function types to be pass-by-reference?
					for i, arg := range e.Args {
						if !fnTyp.GetParameterAccess(i) {
							argTyp := TypeOf(arg, Resolver)
							classify(arg, argTyp)
						}
					}
					// check for the receiver type
					if fnTyp.receiver != nil {
						classify(fnTyp.receiver.Definition(), fnTyp.receiver)
					}
				}
			} else {
				// this is really a type conversion call, keep going
				for _, arg := range e.Args {
					AccessExpr(arg, ReadAccess)
				}
			}
			return
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
		case *ast.LabeledStmt:
			AccessExpr(e.Stmt, ReadAccess)
		case *ast.SliceExpr:
			AccessExpr(e.Low, ReadAccess)
			AccessExpr(e.High, ReadAccess)
			AccessExpr(e.X, t)
		case *ast.DeferStmt:
			AccessExpr(e.Call, ReadAccess)
		case *ast.TypeAssertExpr:
			AccessExpr(e.X, ReadAccess)
		default:
			b.Printf("\x1b[33mUnknown node\x1b[0m %T %+v", e, e)
		}
	}
	AccessExpr(blockNode, ReadAccess)

	if *verbose {
		b.Print("== Defines ==")
		for ident, typ := range dataBlock.defines {
			b.Printf("%s = %v", ident, typ)
		}
		b.Print("== Accesses ==")
		for _, access := range dataBlock.accesses {
			b.Printf(access.String())
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
	return FunctionPassMode
}

func (pass *AccessPass) GetDependencies() []PassType {
	return []PassType{BasicBlockPassType, DefinedTypesPassType}
}

func (pass *AccessPass) RunFunctionPass(fun *ast.FuncDecl, p *Package) (modified bool, err error) {
	pass.ParseBasicBlock(fun, p)
	return
}
