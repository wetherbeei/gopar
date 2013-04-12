// Call Graph
//
// Create a list of the functions every function calls, and check for recursive
// loops.
//
// func A()
// func B()
// func C()

package main

import (
	"fmt"
	"go/ast"
)

type FunctionCallGraph struct {
	pkg   string      // name of the package this function is in
	calls []*FuncType // all of the functions called by this function
}

type CallGraphPass struct {
	BasePass
	funcGraph []*FuncType
}

type CallGraphPassData struct {
	graph map[*FuncType]*FunctionCallGraph // function -> [dependent functions]
}

func NewCallGraphPassData() *CallGraphPassData {
	return &CallGraphPassData{
		graph: make(map[*FuncType]*FunctionCallGraph),
	}
}

func NewCallGraphPass() *CallGraphPass {
	return &CallGraphPass{
		BasePass: NewBasePass(),
	}
}

func (pass *CallGraphPass) GetPassType() PassType {
	return CallGraphPassType
}

func (pass *CallGraphPass) GetPassMode() PassMode {
	return FunctionPassMode
}

func (pass *CallGraphPass) GetDependencies() []PassType {
	return []PassType{DefinedTypesPassType, AccessPassType}
}

type CallGraphPassVisitor struct {
	block    *BasicBlock
	pass     *CallGraphPass
	resolver Resolver
}

func (v *CallGraphPassVisitor) Visit(node ast.Node) (w BasicBlockVisitor) {
	if node != nil {
		switch t := node.(type) {
		case *ast.CallExpr:
			if fnTyp := TypeOf(t.Fun, v.resolver); fnTyp != nil {
				if funcTyp, ok := fnTyp.(*FuncType); ok {
					if *verbose {
						fmt.Println("Found function call to", funcTyp.String())
					}
					v.pass.funcGraph = append(v.pass.funcGraph, funcTyp)
				} else {
					// a type conversion int(x)
					if *verbose {
						fmt.Println("Unknown call", fnTyp.String())
					}
				}
			}
		}
	}

	return v
}

func (v *CallGraphPassVisitor) Done(block *BasicBlock) (bool, error) {
	return false, nil
}

func (pass *CallGraphPass) RunBasicBlockPass(block *BasicBlock, p *Package) BasicBlockVisitor {
	return &CallGraphPassVisitor{block: block, pass: pass, resolver: MakeResolver(block, p, pass.compiler)}
}

func (pass *CallGraphPass) RunFunctionPass(fun *ast.FuncDecl, p *Package) (modified bool, err error) {

	callGraph, ok := pass.GetResult(p).(*CallGraphPassData)
	if !ok {
		callGraph = NewCallGraphPassData()
		pass.SetResult(p, callGraph)
	}

	block := pass.GetCompiler().GetPassResult(BasicBlockPassType, fun).(*BasicBlock)
	modified, err = RunBasicBlock(pass, block, p)
	resolver := MakeResolver(nil, p, pass.compiler)

	var fnTyp *FuncType
	name := fun.Name.Name
	if fun.Recv != nil {
		recvTyp := TypeOfDecl(fun.Recv.List[0].Type, resolver)
		fnTyp = recvTyp.Method(name)
	} else {
		fnTyp = resolver(name).(*FuncType)
	}
	if *verbose {
		block.Print("Call graph", fnTyp.String())
	}
	callGraph.graph[fnTyp] = &FunctionCallGraph{pkg: p.name, calls: pass.funcGraph}
	if *verbose {
		fmt.Println(fun.Name.Name, pass.funcGraph)
	}
	pass.funcGraph = nil
	return
}
