// Access pass function propogation
//
// Alias accesses to function parameters passed by pointers, and record
// global variable accesses. Propogate accesses upwards through the blocks.
//
// func foo(ptr *DataA, val DataB) {
//   // Read ptr
//   // Write ptr
//   // Read val
//   // Write val
// }
// 
// func bar(index int, val []DataB) int {
//   return val[index]
// }
//
// func main() {
//   a := &DataA{}
//   foo(a, DataB{}) // bubble up the reads and writes to "a" (ptr)
//   i := 0
//   aList := []DataA{a}
//   bar(i, aList) // bubble up val[index] -> aList[i] access
// }
package main

import (
	"go/ast"
)

type AccessPassFuncPropogate struct {
	BasePass
}

type AccessPassFuncPropogateVisitor struct {
	pass      Pass
	p         *Package
	cur       *BasicBlock
	dataBlock *AccessPassData
	node      ast.Node
}

func (v AccessPassFuncPropogateVisitor) Done(block *BasicBlock) (modified bool, err error) {
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

func (v AccessPassFuncPropogateVisitor) Visit(node ast.Node) (w BasicBlockVisitor) {
	// Get the closest enclosing basic block for this node
	dataBlock := v.dataBlock
	b := v.cur

	if node == nil {
		// post-order actions (all sub-nodes have been visited)
		node = v.node
		// Locate function calls
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
				funcDataBlock := pass.compiler.GetPassResult(BasicBlockPassType, fnDecl).(*BasicBlock).Get(AccessPassType).(*AccessPassData)

				child := v.cur
				parent := child.parent
				// Now fill in the accesses this call would have made, and propogate it
				// all the way to the top
				var funcAccesses []IdentifierGroup // only the accesses this function made

				// Fill in global accesses
				for _, access := range funcDataBlock.accesses {
					if _, ok := funcDataBlock.defines[access.group[0].id]; !ok {
						// if there is an array access that uses an identifier block defined in 
						// this block, change the access from b[idx] to b
						var ig IdentifierGroup = access
						for idx, ident := range access.group {
							if _, ok := dataBlock.defines[ident.index]; ok && ident.isIndexed {
								ig.group = make([]Identifier, idx+1)
								copy(ig.group, access.group)
								parent.Printf("Leaving index scope [%s]", ig.group[idx].index)
								ig.group[idx].isIndexed = false
								ig.group[idx].index = ""
							}
							break
						}
						funcAccesses = append(funcAccesses, ig)
					}
				}

				// Fill in aliased arguments
				pos := 0 // argument position
				for _, arg := range funcType.Params.List {
					writeThrough := false
					// is the argument able to be modified?
					// builtin types (slice, map, chan), pointers
					switch arg.Type.(type) {
					case *ast.ArrayType, *ast.MapType, *ast.ChanType:
						b.Printf("Pass-by-reference %v %T", arg.Names, arg.Type)
						writeThrough = true
					case *ast.StarExpr:
						b.Printf("Pass-by-pointer %v %T", arg.Names, arg.Type)
						writeThrough = true
					}

					for _, argName := range arg.Names {
						if writeThrough {
							callArg := t.Args[pos]
							callIdent := &IdentifierGroup{}
							AccessIdentBuild(callIdent, callArg)

							// Find all accesses to these variables
							for _, access := range funcDataBlock.accesses {
								// Replace the function arg name with the callIdent prefix
								if access.group[0].id == argName.Name {

								}
							}
							b.Print(callArg, argName)
						}
						pos++
					}
				}

				for _, a := range dataBlock.accesses {
					b.Print(a.String())
				}
				// Propogate ONLY aliased argument accesses upwards
				// NOTE: doesn't work with recursive functions

				// Move upwards, replacing the placeholder access with the group of
				// accesses made by this function. Stop at variable define boundaries
				placeholderIdent := v.pass.GetCompiler().GetPassResult(AccessPassType, t).(*ast.Ident)
				b.Printf("\x1b[33m>> %s\x1b[0m filling in function effects: %+v, %+v", placeholderIdent.Name, t, funcDecl)

				// Walk up the parent blocks
				for ; child != nil; child = child.parent {
					// Find the placeholder
					dataBlock := child.Get(AccessPassType).(*AccessPassData)
					var idx int
					var val IdentifierGroup
					for idx, val = range dataBlock.accesses {
						if val.group[0].id == placeholderIdent.Name {
							break
						}
					}
					b.Print(idx, dataBlock)
					// Remove the placeholder, insert the function accesses
					// TODO....
				}
				return nil
			}
		}
		return v
	}

	b.Printf("start %T %+v", node, node)
	v.node = node

	return v
}

func NewAccessPassFuncPropogate() *AccessPassFuncPropogate {
	return &AccessPassFuncPropogate{
		BasePass: NewBasePass(),
	}
}

func (pass *AccessPassFuncPropogate) GetPassType() PassType {
	return AccessPassFuncPropogateType
}

func (pass *AccessPassFuncPropogate) GetPassMode() PassMode {
	return ModulePassMode
}

func (pass *AccessPassFuncPropogate) GetDependencies() []PassType {
	return []PassType{CallGraphPassType, AccessPassPropogateType}
}

// Declare two Run* functions

func (pass *AccessPassFuncPropogate) RunBasicBlockPass(block *BasicBlock, p *Package) BasicBlockVisitor {
	dataBlock := block.Get(AccessPassType).(*AccessPassData)
	return AccessPassFuncPropogateVisitor{pass: pass, cur: block, dataBlock: dataBlock, p: p}
}

func (pass *AccessPassFuncPropogate) RunModulePass(file *ast.File, p *Package) (modified bool, err error) {
	callGraph := pass.compiler.GetPassResult(CallGraphPassType, nil).(*CallGraphPassData)
	run := make(map[string]bool) // which functions have been propogated
	var runOrder []string
	for fn, callDepends := range callGraph.graph {
		runOrder = append(runOrder, fn)
		runOrder = append(runOrder, callDepends...)
	}

	for _, fnName := range runOrder {
		fn := p.Lookup(fnName)
		if fn == nil || run[fnName] {
			continue
		}
		fnDecl := fn.Decl.(*ast.FuncDecl)
		block := pass.compiler.GetPassResult(BasicBlockPassType, fnDecl).(*BasicBlock)

		// Manually run the basic block pass in inverse call graph order
		var mod bool
		mod, err = RunBasicBlock(pass, block, p)
		modified = modified || mod
		if err != nil {
			return
		}
		run[fnName] = true
	}
	return
}
