// Access pass propogation
//
// Alias accesses to function parameters passed by pointers, and record
// global variable accesses. Propogate accesses upwards through the blocks.
//
//
// func foo(ptr *DataA, val DataB) {
//   // Read ptr
//   // Write ptr
//   // Read val
//   // Write val
// }
// 
// func main() {
//   a := &DataA{}
//   foo(a, DataB{}) // bubble up the reads and writes to "a" (ptr)
// }
package main

import (
	"go/ast"
)

type AccessPassPropogate struct {
	BasePass
}

type AccessPassPropogateVisitor struct {
	p         *Package
	cur       *BasicBlock
	dataBlock *AccessPassData
}

func (v AccessPassPropogateVisitor) Done(block *BasicBlock) (modified bool, err error) {
	dataBlock := v.dataBlock

	block.Print("== Defines ==")
	for ident, expr := range dataBlock.defines {
		block.Printf("%s = %T %+v", ident, expr, expr)
	}
	block.Print("== Accesses ==")
	for _, access := range dataBlock.accesses {
		block.Printf(access.String())
	}

	MergeDependenciesUpwards(block)
	return
}

func (v AccessPassPropogateVisitor) Visit(node ast.Node) (w BasicBlockVisitor) {
	// Get the closest enclosing basic block for this node
	dataBlock := v.dataBlock
	b := v.cur

	if node == nil {
		// post-order actions (all sub-nodes have been visited)
		return v
	}

	b.Printf("start %T %+v", node, node)

	// Locate 
	switch t := node.(type) {
	case *ast.CallExpr:
		switch f := t.Fun.(type) {
		case *ast.FuncLit:
			// go down a FuncLit branch
		case *ast.Ident:
			fun := v.p.Lookup(f.Name)
			if fun == nil {
				// builtin function, or not found
				b.Print("Function not found", f.Name)
				return nil
			}
			funcDecl := fun.Decl.(*ast.FuncDecl)
			funcType := funcDecl.Type
			b.Printf("\x1b[33m+\x1b[0m filling in function effects: %+v, %+v", t, funcDecl)

			// Now fill in the accesses this call would have made
			pos := 0 // argument position
			for _, arg := range funcType.Params.List {
				// is the argument able to be modified?
				// builtin types (slice, map, chan), pointers
				writeThrough := false
				switch arg.Type.(type) {
				case *ast.ArrayType, *ast.MapType, *ast.ChanType:
					b.Printf("Pass-by-reference %T", arg.Type)
					writeThrough = true
				case *ast.StarExpr:
					b.Printf("Pass-by-pointer %T", arg.Type)
					writeThrough = true
				}
				for _, name := range arg.Names {
					callArg := t.Args[pos]
					b.Print(callArg, name, writeThrough)
					pos++
				}
			}

			for _, a := range dataBlock.accesses {
				b.Print(a.String())
			}
			//v.pass.SetResult(t, b)
			//AccessExpr(f, ReadAccess)
			return nil
		}
	}
	return v
}

func NewAccessPassPropogate() *AccessPassPropogate {
	return &AccessPassPropogate{
		BasePass: NewBasePass(),
	}
}

func (pass *AccessPassPropogate) GetPassType() PassType {
	return AccessPassPropogateType
}

func (pass *AccessPassPropogate) GetPassMode() PassMode {
	return BasicBlockPassMode
}

func (pass *AccessPassPropogate) GetDependencies() []PassType {
	return []PassType{AccessPassType}
}

func MergeDependenciesUpwards(child *BasicBlock) {
	// TODO: merge reads/writes of identifiers outside this scope
	if child.parent == nil {
		return
	}
	parent := child.parent
	dataBlock := child.Get(AccessPassType).(*AccessPassData)
	parentDataBlock := parent.Get(AccessPassType).(*AccessPassData)
	for _, access := range dataBlock.accesses {
		// move to parent if not defined in this block
		if _, ok := dataBlock.defines[access.group[0].id]; !ok {
			var ig IdentifierGroup = access
			parentDataBlock.accesses = append(parentDataBlock.accesses, ig)
			parent.Print("<< Merged upwards", ig.String())
		}
	}
	return
}

func (pass *AccessPassPropogate) RunBasicBlockPass(block *BasicBlock, p *Package) BasicBlockVisitor {
	dataBlock := block.Get(AccessPassType).(*AccessPassData)

	return AccessPassPropogateVisitor{cur: block, dataBlock: dataBlock, p: p}
}
