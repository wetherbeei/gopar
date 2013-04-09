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
	"fmt"
	"go/ast"
)

type AccessPassFuncPropogate struct {
	BasePass
}

func StripDefinedIndex(access IdentifierGroup, data *AccessPassData) (ig IdentifierGroup) {
	// check if an index variable is also a function argument and
	// remove it
	ig = access
	newAccess := make([]Identifier, len(access.group))

	copy(newAccess, access.group) // full copy
	for idx, ident := range newAccess {
		if _, ok := data.defines[ident.index]; ok && ident.isIndexed {
			newAccess = newAccess[0 : idx+1]
			newAccess[idx].isIndexed = false
			newAccess[idx].index = ""
		}
		break
	}
	ig.group = newAccess
	return
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

	return
}

// TODO: this is one big function, split it up
func (v AccessPassFuncPropogateVisitor) Visit(node ast.Node) (w BasicBlockVisitor) {

	if node != nil {
		// decend into this node
		v.node = node
		return v
	}
	// post-order actions (all sub-nodes have been visited)
	node = v.node

	// Locate function calls
	// Get the closest enclosing basic block for this node
	b := v.cur
	b.Print(node)
	pass := v.pass
	resolver := MakeResolver(b, v.p, v.pass.GetCompiler())

	var fun *FuncType
	var call *ast.CallExpr
	switch t := node.(type) {
	case *ast.CallExpr:

		if fnTyp := TypeOf(t.Fun, resolver); fnTyp != nil {
			if funcTyp, ok := fnTyp.(*FuncType); ok {
				if funcTyp.body == nil {
					// functions without definitions are assumed to modify all arguments
					// in access pass
					return v
				}
				fun = funcTyp
				call = t
			} else {
				// not a call, it's a type cast
				return v
			}
			b.Print("Found function call", fnTyp.String())
		} else {
			// we already recorded writes for all of the arguments in access pass
			return v
		}
	default:
		return v
	}
	b.Printf("%T %+v %s", fun.Definition(), fun.Definition(), fun.name)
	funcType := fun.typ
	funcDataBlock := pass.GetCompiler().GetPassResult(BasicBlockPassType, fun.Definition()).(*BasicBlock).Get(AccessPassType).(*AccessPassData)

	// Now fill in the accesses this call would have made, and propogate it
	// all the way to the top
	var funcAccesses []IdentifierGroup // only the accesses this function made

	// Fill in global accesses
	for _, access := range funcDataBlock.accesses {
		if _, ok := funcDataBlock.defines[access.group[0].id]; !ok {
			// if there is an array access that uses an identifier block defined in 
			// this block, change the access from b[idx] to b
			ig := StripDefinedIndex(access, funcDataBlock)
			b.Print("Global access", ig.String())
			funcAccesses = append(funcAccesses, ig)
		}
	}

	// copy matching accesses from arg/argName to callArg
	// func foo(arg) {}
	// foo(callArg) -> translate accesses inside foo to "arg" to "callArg"s
	// CallExpr site
	propogateFn := func(callArg ast.Expr, arg *ast.Field, argName *ast.Ident) {
		callIdent := &IdentifierGroup{}
		AccessIdentBuild(callIdent, callArg, nil)

		// Find all accesses to these variables
		for _, access := range funcDataBlock.accesses {
			// Replace the function arg name with the callIdent prefix
			if access.group[0].id == argName.Name {
				// check if an index variable is also a function argument and
				// remove it
				ig := StripDefinedIndex(access, funcDataBlock)

				newAccess := ig.group
				// if the callsite is &a and the access is *a, make the access
				// a for this function
				var callIdentCopy []Identifier
				if callIdent.group[len(callIdent.group)-1].refType == AddressOf && newAccess[len(newAccess)-1].refType == Dereference {
					b.Print("Removing pointer alias & -> *")
					newAccess[len(newAccess)-1].refType = NoReference
					callIdentCopy = make([]Identifier, len(callIdent.group))
					copy(callIdentCopy, callIdent.group)
					callIdentCopy[len(callIdentCopy)-1].refType = NoReference
				} else {
					callIdentCopy = callIdent.group
				}
				// replace access[0] with callIdent
				ig.group = append(ig.group, callIdentCopy...)
				ig.group = append(ig.group, newAccess[1:]...)
				b.Printf("%s -> %s", access.String(), ig.String())
				funcAccesses = append(funcAccesses, ig)
			}
		}
	}

	// Fill in aliased arguments
	pos := 0 // argument position
	for _, arg := range funcType.Params.List {
		writeThrough := !TypeOfDecl(arg.Type, resolver).PassByValue()
		// is the argument able to be modified?
		// builtin types (slice, map, chan), pointers

		for _, argName := range arg.Names {
			callArg := call.Args[pos]
			if writeThrough {
				propogateFn(callArg, arg, argName)
			}
			pos++
		}
	}

	// also propogate accesses to the receiver
	if fun.receiver != nil {
		if !fun.receiver.PassByValue() {
			recv := fun.Node.(*ast.FuncDecl).Recv.List[0]
			// callArg is the struct we're calling this method on
			callArg := call.Fun.(*ast.SelectorExpr).X
			propogateFn(callArg, recv, recv.Names[0])
		}
	}

	// Propogate ONLY aliased argument accesses upwards
	// NOTE: doesn't work with recursive functions??

	// Move upwards, replacing the placeholder access with the group of
	// accesses made by this function. Stop at variable define boundaries
	placeholderIdent := v.pass.GetCompiler().GetPassResult(AccessPassType, call).(*ast.Ident)
	b.Printf("\x1b[33m>> %s\x1b[0m filling in function effects: %+v, %s", placeholderIdent.Name, call, fun.String())

	// Walk up the parent blocks
	child := b
	for ; child != nil; child = child.parent {
		// Find the placeholder
		dataBlock := child.Get(AccessPassType).(*AccessPassData)
		var placeholderIdx int
		var val IdentifierGroup
		for placeholderIdx, val = range dataBlock.accesses {
			if val.group[0].id == placeholderIdent.Name {
				break
			}
		}
		b.Printf("Replacing placeholder at %d", placeholderIdx)

		// Remove the placeholder
		dataBlock.accesses = append(dataBlock.accesses[0:placeholderIdx], dataBlock.accesses[placeholderIdx+1:]...)
		// Insert the function accesses, do a deep copy
		var funcAccessCopy []IdentifierGroup
		for _, v := range funcAccesses {
			// deep copy the identifers
			groupCopy := make([]Identifier, len(v.group))
			copy(groupCopy, v.group)
			v.group = groupCopy
			funcAccessCopy = append(funcAccessCopy, v)
		}

		b.Print(" << Propogating up")
		for _, a := range funcAccessCopy {
			b.Print(a.String())
		}
		dataBlock.accesses = append(dataBlock.accesses[0:placeholderIdx], append(funcAccessCopy, dataBlock.accesses[placeholderIdx:]...)...)

		// Check if the identifier leaves scope
		for idx := 0; idx < len(funcAccesses); {
			access := funcAccesses[idx]
			if _, ok := dataBlock.defines[access.group[0].id]; ok {
				b.Print("Leaving scope", access.String())
				funcAccesses = append(funcAccesses[:idx], funcAccesses[idx+1:]...)
			} else {
				idx++
			}
		}

		// Check if an index variable leaves scope
		for accIdx, access := range funcAccesses {
			// check if an index variable is also a function argument and
			// remove it
			funcAccesses[accIdx] = StripDefinedIndex(access, dataBlock)
		}
	}
	return nil
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
	callGraph := pass.compiler.GetPassResult(CallGraphPassType, p).(*CallGraphPassData)
	run := make(map[*FuncType]bool) // which functions have been propogated
	var orderGraph func(map[*FuncType][]*FuncType, *FuncType) []*FuncType
	added := make(map[*FuncType]bool)
	orderGraph = func(graph map[*FuncType][]*FuncType, f *FuncType) (result []*FuncType) {
		fmt.Println(f.String())
		for _, fn := range callGraph.graph[f] {
			if !added[fn] {
				added[fn] = true // prevent a recursive loop
				result = append(result, orderGraph(graph, fn)...)
			}
		}
		result = append(result, f)
		return
	}

	// spit out every function in the call graph with all of their dependencies
	// listed before them...you can start from the "main" function for the main
	// package, but supporting packages don't have a single entry point
	var runOrder []*FuncType
	for k, _ := range callGraph.graph {
		fnOrder := orderGraph(callGraph.graph, k)
		runOrder = append(runOrder, fnOrder...)
	}
	for _, fnDecl := range runOrder {
		if fnDecl == nil || run[fnDecl] {
			continue
		}
		if fnDecl.body == nil {
			continue
		}
		block := pass.compiler.GetPassResult(BasicBlockPassType, fnDecl.body).(*BasicBlock)

		// Manually run the basic block pass in inverse call graph order
		var mod bool
		mod, err = RunBasicBlock(pass, block, p)
		modified = modified || mod
		if err != nil {
			return
		}
		run[fnDecl] = true
	}
	return
}
