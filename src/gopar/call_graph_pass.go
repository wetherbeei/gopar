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

type CallGraphPass struct {
	BasePass
}

type CallGraphPassData struct {
	graph map[*ast.BlockStmt][]*ast.BlockStmt // function -> [dependent functions]
}

func NewCallGraphPassData() *CallGraphPassData {
	return &CallGraphPassData{
		graph: make(map[*ast.BlockStmt][]*ast.BlockStmt),
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
	return []PassType{DefinedTypesPassType}
}

func (pass *CallGraphPass) RunFunctionPass(fun *ast.FuncDecl, p *Package) (modified bool, err error) {
	callGraph, ok := pass.GetResult(p).(*CallGraphPassData)
	if !ok {
		callGraph = NewCallGraphPassData()
		pass.SetResult(p, callGraph)
	}

	var external []*ast.BlockStmt
	resolver := MakeResolver(nil, p, pass.compiler)
	ast.Inspect(fun, func(node ast.Node) bool {
		if node != nil {
			switch t := node.(type) {
			case *ast.CallExpr:
				fnTyp := TypeOf(t, resolver)
				if fnTyp != nil {
					external = append(external, fnTyp.(*FuncType).body)
				}
			}
			return true
		}
		return false
	})
	callGraph.graph[fun.Body] = external
	fmt.Println(fun.Name.Name, external)
	return
}
