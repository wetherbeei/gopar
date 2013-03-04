// Parallelize pass (finally)
//
// Range statements must be made up of only:
// ReadAccess for any variable
// WriteAccess to any variable indexed by the range index
// WriteFirst access to any variable (make a private copy)
// - for privatizing, copy the last iteration's value into a variable
//     i = 0
//     for i, v = range array {
//     }
//     i and v should equal 

package main

import (
//"go/ast"
)

type ParallelizePass struct {
	BasePass
}

type ParallelizeData struct {
	// tag with map[ast.RangeStmt]s = some data
}

func NewParallelizeData() *ParallelizeData {
	return nil
}

func NewParallelizePass() *ParallelizePass {
	return &ParallelizePass{
		BasePass: NewBasePass(),
	}
}

func (pass *ParallelizePass) GetPassType() PassType {
	return ParallelizePassType
}

func (pass *ParallelizePass) GetPassMode() PassMode {
	return BasicBlockPassMode
}

func (pass *ParallelizePass) GetDependencies() []PassType {
	return []PassType{DependencyPassType}
}

func (pass *ParallelizePass) RunBasicBlockPass(block *BasicBlock, p *Package) BasicBlockVisitor {
	dependencyData := block.Get(DependencyPassType).(*DependencyPassData)
	//data := NewParallelizeData()

	block.Print("== Dependencies ==")
	for _, dep := range dependencyData.deps {
		block.Print(dep.String())
	}

	return DefaultBasicBlockVisitor{}
}
