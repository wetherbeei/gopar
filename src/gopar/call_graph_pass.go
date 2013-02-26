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
	graph map[string][]string // inverted map from 
}

func NewCallGraphPassData() *CallGraphPassData {
	return &CallGraphPassData{
		graph: make(map[string][]string),
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
	return []PassType{}
}

func (pass *CallGraphPass) RunFunctionPass(fun *ast.FuncDecl, p *Package) (modified bool, err error) {
	callGraph, ok := pass.GetResult(nil).(*CallGraphPassData)
	if !ok {
		callGraph = NewCallGraphPassData()
		pass.SetResult(nil, callGraph)
	}
	var fnName = fun.Name.Name
	var external []string
	ast.Inspect(fun, func(node ast.Node) bool {
		if node != nil {
			switch t := node.(type) {
			case *ast.CallExpr:
				switch f := t.Fun.(type) {
				case *ast.Ident:
					var name string = f.Name
					external = append(external, name)
				}
			}
			return true
		}
		return false
	})
	callGraph.graph[fnName] = external
	fmt.Println(fnName, external)
	return
}
